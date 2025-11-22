package mockserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"
)

// MockServer представляет мок-сервер для внешних клиентов
type MockServer struct {
	router   *chi.Mux
	server   *httptest.Server
	handlers []*Handler
	mu       sync.RWMutex
}

// Handler представляет обработчик для конкретного эндпоинта
type Handler struct {
	method             string
	url                string
	expBody            string
	actualBody         string
	respPath           string
	statusCode         int
	calledCh           chan struct{}
	checkRequest       func([]byte)
	checkRequestHeader func(header http.Header)
	skipCalled         bool
}

// NewMockServer создает новый мок-сервер
func NewMockServer() *MockServer {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(60 * time.Second))

	ms := &MockServer{
		router:   router,
		handlers: make([]*Handler, 0),
	}

	// Настраиваем универсальный обработчик для всех запросов
	router.HandleFunc("/*", ms.handleRequest)

	// Используем httptest.Server для автоматического выбора порта
	ms.server = httptest.NewServer(router)

	return ms
}

// handleRequest обрабатывает все входящие запросы
func (ms *MockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Ищем первый неиспользованный обработчик
	for i, h := range ms.handlers {
		// Проверяем, что обработчик еще не был вызван
		select {
		case <-h.calledCh:
			// Обработчик уже был использован, пропускаем
			continue
		default:
		}

		if h.method == r.Method && h.url == r.URL.Path {
			// Читаем тело запроса
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusInternalServerError)
				return
			}
			h.actualBody = string(body)

			// Выполняем кастомную проверку тела запроса
			if h.checkRequest != nil {
				h.checkRequest(body)
			}

			// Выполняем кастомную проверку заголовков
			if h.checkRequestHeader != nil {
				h.checkRequestHeader(r.Header)
			}

			// Загружаем ответ из файла
			if h.respPath != "" {
				basePath := getSamplesPath()
				respBody, err := os.ReadFile(path.Join(basePath, h.respPath))
				if err != nil {
					http.Error(w, fmt.Sprintf("failed to load response file: %v", err), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(h.statusCode)
				w.Write(respBody)
			} else {
				w.WriteHeader(h.statusCode)
			}

			// Сигнализируем о вызове
			h.calledCh <- struct{}{}

			// Удаляем использованный обработчик из списка
			ms.handlers = append(ms.handlers[:i], ms.handlers[i+1:]...)

			return
		}
	}

	// Обработчик не найден
	http.Error(w, fmt.Sprintf("no handler found for %s %s", r.Method, r.URL.Path), http.StatusNotFound)
}

// Expect начинает настройку ожидания для эндпоинта
func (ms *MockServer) Expect(method, url string) *Handler {
	h := &Handler{
		method:     method,
		url:        url,
		statusCode: http.StatusOK,
		calledCh:   make(chan struct{}, 1),
	}

	ms.mu.Lock()
	ms.handlers = append(ms.handlers, h)
	ms.mu.Unlock()

	return h
}

// Return устанавливает путь к файлу с ответом
func (h *Handler) Return(respPath string) *Handler {
	h.respPath = respPath
	return h
}

// HTTPCode устанавливает HTTP статус код
func (h *Handler) HTTPCode(statusCode int) *Handler {
	h.statusCode = statusCode
	return h
}

// RequireBody устанавливает ожидаемое тело запроса из файла
func (h *Handler) RequireBody(bodyPath string) *Handler {
	basePath := getSamplesPath()
	body, err := os.ReadFile(path.Join(basePath, bodyPath))
	if err != nil {
		panic(fmt.Sprintf("failed to load expected body file: %v", err))
	}
	h.expBody = string(body)
	return h
}

// CheckRequest устанавливает кастомную функцию проверки тела запроса
func (h *Handler) CheckRequest(fn func([]byte)) *Handler {
	h.checkRequest = fn
	return h
}

// CheckHeader устанавливает кастомную функцию проверки заголовков
func (h *Handler) CheckHeader(fn func(header http.Header)) *Handler {
	h.checkRequestHeader = fn
	return h
}

// SkipCalled пропускает проверку вызова эндпоинта
func (h *Handler) SkipCalled() *Handler {
	h.skipCalled = true
	return h
}

// Clear очищает все обработчики и проверяет, что все были вызваны
func (ms *MockServer) Clear(t *testing.T) {
	ms.mu.Lock()
	defer func() {
		ms.handlers = nil
		ms.mu.Unlock()
	}()

	timer := time.NewTimer(1200 * time.Millisecond)
	defer timer.Stop()

	for _, h := range ms.handlers {
		if h.skipCalled {
			continue
		}

		select {
		case <-h.calledCh:
			// Эндпоинт был вызван - проверяем тело если нужно
			if h.expBody != "" {
				require.JSONEq(t, h.expBody, h.actualBody, "request body mismatch for %s %s", h.method, h.url)
			}
		case <-timer.C:
			// Тайм-аут - эндпоинт НЕ был вызван
			require.FailNow(t, "timeout exceeded", "expected call to %s %s", h.method, h.url)
			return
		}
	}
}

// Stop останавливает мок-сервер
func (ms *MockServer) Stop() {
	if ms.server != nil {
		ms.server.Close()
	}
}

// GetBaseURL возвращает базовый URL мок-сервера
func (ms *MockServer) GetBaseURL() string {
	return ms.server.URL
}

// getSamplesPath возвращает путь к директории с samples
func getSamplesPath() string {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	return filepath.Join(dir, "..", "samples")
}

// respondWithError отправляет ответ с ошибкой
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
	})
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

// GetSampleData загружает данные из файла sample
func GetSampleData(samplePath string) []byte {
	basePath := getSamplesPath()
	fileBody, err := os.ReadFile(path.Join(basePath, samplePath))
	if err != nil {
		panic(fmt.Sprintf("failed to load sample file `%s`: %s", samplePath, err))
	}
	return fileBody
}

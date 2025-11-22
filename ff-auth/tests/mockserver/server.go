package mockserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// MockServer представляет мок-сервер для внешних клиентов
type MockServer struct {
	router   *chi.Mux
	server   *httptest.Server
	handlers *MockHandlers
}

// MockHandlers содержит обработчики для различных сценариев
type MockHandlers struct {
	// RegisterUserFromService определяет поведение для регистрации пользователя
	RegisterUserFromService func(w http.ResponseWriter, r *http.Request)
}

// NewMockServer создает новый мок-сервер
func NewMockServer() *MockServer {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Timeout(60 * time.Second))

	handlers := &MockHandlers{
		RegisterUserFromService: defaultRegisterUserFromService,
	}

	ms := &MockServer{
		router:   router,
		handlers: handlers,
	}

	ms.setupRoutes()

	// Используем httptest.Server для автоматического выбора порта
	ms.server = httptest.NewServer(router)

	return ms
}

// setupRoutes настраивает маршруты мок-сервера
func (ms *MockServer) setupRoutes() {
	ms.router.Route("/api/v1", func(r chi.Router) {
		r.Route("/internal", func(r chi.Router) {
			// Используем обертку, которая вызывает функцию из handlers при каждом запросе
			r.Post("/users/register", func(w http.ResponseWriter, r *http.Request) {
				ms.handlers.RegisterUserFromService(w, r)
			})
		})
	})
}

// Start запускает мок-сервер (для httptest.Server не требуется)
func (ms *MockServer) Start() error {
	// httptest.Server запускается автоматически при создании
	return nil
}

// Stop останавливает мок-сервер
func (ms *MockServer) Stop() {
	ms.server.Close()
}

// SetRegisterUserFromServiceHandler устанавливает кастомный обработчик для регистрации пользователя
func (ms *MockServer) SetRegisterUserFromServiceHandler(handler func(w http.ResponseWriter, r *http.Request)) {
	ms.handlers.RegisterUserFromService = handler
}

// GetBaseURL возвращает базовый URL мок-сервера
func (ms *MockServer) GetBaseURL() string {
	return ms.server.URL
}

// defaultRegisterUserFromService - обработчик по умолчанию для регистрации пользователя
func defaultRegisterUserFromService(w http.ResponseWriter, r *http.Request) {
	var req ServiceRegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "некорректный формат запроса")
		return
	}

	// Валидация запроса
	if req.UserID == 0 {
		respondWithError(w, http.StatusBadRequest, "user_id обязателен")
		return
	}
	if req.Email == "" {
		respondWithError(w, http.StatusBadRequest, "email обязателен")
		return
	}
	if req.Nickname == "" {
		respondWithError(w, http.StatusBadRequest, "nickname обязателен")
		return
	}

	// Успешный ответ
	now := time.Now().Unix()
	response := UserDTO{
		ID:        req.UserID,
		Email:     req.Email,
		Nickname:  req.Nickname,
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// respondWithError отправляет ответ с ошибкой
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
	})
}

// ServiceRegisterUserRequest представляет запрос на регистрацию пользователя от сервиса
type ServiceRegisterUserRequest struct {
	UserID   int64   `json:"user_id"`
	Email    string  `json:"email"`
	Nickname string  `json:"nickname"`
	Name     *string `json:"name,omitempty"`
}

// UserDTO представляет данные пользователя
type UserDTO struct {
	ID        int64   `json:"id"`
	Email     string  `json:"email"`
	Nickname  string  `json:"nickname"`
	Name      *string `json:"name,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	AvatarID  *string `json:"avatar_id,omitempty"`
	Birthdate *int64  `json:"birthdate,omitempty"`
	CreatedAt int64   `json:"created_at"`
	UpdatedAt int64   `json:"updated_at"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
}

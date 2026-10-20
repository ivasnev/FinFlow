package container

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-files/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-files/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
	tvmmiddleware "github.com/ivasnev/FinFlow/ff-tvm/pkg/middleware"
)

// protectedStorage фиксирует обращение к хранилищу после авторизации.
type protectedStorage struct {
	service.MinIO
	deleted int
}

// DeleteOne фиксирует удаление в тестовом хранилище.
func (s *protectedStorage) DeleteOne(string) error {
	s.deleted++
	return nil
}

// publicKeys предоставляет ключ подписавшего билет сервиса.
type publicKeys struct{ key string }

// GetPublicKey возвращает публичный ключ отправителя.
func (k publicKeys) GetPublicKey(int) (string, error) { return k.key, nil }

// TestFileRoutesRequireTicket проверяет штатную регистрацию всех файловых маршрутов.
func TestFileRoutesRequireTicket(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.TVM.ServiceID = 3
	storage := &protectedStorage{}
	c := &Container{Config: cfg, Router: gin.New(), ServerHandler: handler.NewServerHandler(storage)}
	c.RegisterRoutes()
	for _, tc := range []struct{ method, path string }{
		{"POST", "/api/v1/files"},
		{"POST", "/api/v1/files/many"},
		{"GET", "/api/v1/files/many"},
		{"DELETE", "/api/v1/files/many"},
		{"GET", "/api/v1/files/11111111-1111-4111-8111-111111111111"},
		{"DELETE", "/api/v1/files/11111111-1111-4111-8111-111111111111"},
		{"GET", "/api/v1/files/11111111-1111-4111-8111-111111111111/metadata"},
		{"POST", "/api/v1/files/11111111-1111-4111-8111-111111111111/temporary-url"},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			res := httptest.NewRecorder()
			c.Router.ServeHTTP(res, httptest.NewRequest(tc.method, tc.path, nil))
			if res.Code != http.StatusUnauthorized || storage.deleted != 0 {
				t.Fatalf("status=%d deleted=%d body=%s", res.Code, storage.deleted, res.Body)
			}
		})
	}
}

// TestFileTicketRecipient проверяет передачу ID получателя из конфигурации в TVM middleware.
func TestFileTicketRecipient(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, recipient := range []int{3, 2} {
		t.Run(strconv.Itoa(recipient), func(t *testing.T) {
			cfg := &config.Config{}
			cfg.TVM.ServiceID = 3
			storage := &protectedStorage{}
			c := &Container{Config: cfg, Router: gin.New(), TVMClient: publicKeys{base64.StdEncoding.EncodeToString(pub)}, ServerHandler: handler.NewServerHandler(storage)}
			c.RegisterRoutes()
			payload := struct {
				From     int    `json:"from"`
				To       int    `json:"to"`
				TTL      int64  `json:"ttl"`
				Metadata string `json:"metadata"`
			}{2, recipient, time.Now().Add(time.Hour).Unix(), "{}"}
			data, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			ticket, err := json.Marshal(struct {
				Payload   interface{} `json:"payload"`
				Signature string      `json:"signature"`
			}{payload, base64.StdEncoding.EncodeToString(ed25519.Sign(priv, data))})
			if err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest("DELETE", "/api/v1/files/11111111-1111-4111-8111-111111111111", nil)
			req.Header.Set(tvmmiddleware.HeaderServiceTicket, "serv:Mg==:"+base64.StdEncoding.EncodeToString(ticket))
			res := httptest.NewRecorder()
			c.Router.ServeHTTP(res, req)
			if recipient == 3 {
				if res.Code != 200 || storage.deleted != 1 {
					t.Fatalf("valid ticket: status=%d deleted=%d body=%s", res.Code, storage.deleted, res.Body)
				}
			} else if res.Code != 401 || storage.deleted != 0 {
				t.Fatalf("foreign ticket: status=%d deleted=%d", res.Code, storage.deleted)
			}
		})
	}
}

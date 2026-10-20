package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/service"
)

// adminRepository фиксирует вызовы административных операций в HTTP-тесте.
type adminRepository struct {
	service.ServiceRepository
	calls int
}

// Create сохраняет факт регистрации сервиса.
func (r *adminRepository) Create(_ context.Context, svc *service.Service) error {
	r.calls++
	svc.ID = 1
	return nil
}

// GrantAccess сохраняет факт выдачи доступа.
func (r *adminRepository) GrantAccess(context.Context, int, int) error {
	r.calls++
	return nil
}

// RevokeAccess сохраняет факт отзыва доступа.
func (r *adminRepository) RevokeAccess(context.Context, int, int) error {
	r.calls++
	return nil
}

// TestAdminRoutes проверяет границу авторизации на настоящих маршрутах Gin.
func TestAdminRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, path := range []string{"/service", "/access/grant", "/access/revoke"} {
		for _, tc := range []struct {
			name, configured, header string
			allowed                  bool
		}{
			{"missing", "admin-secret", "", false},
			{"wrong", "admin-secret", "Bearer wrong", false},
			{"raw", "admin-secret", "admin-secret", false},
			{"unset", "", "Bearer ", false},
			{"unsetWithToken", "", "Bearer admin-secret", false},
			{"valid", "admin-secret", "Bearer admin-secret", true},
		} {
			t.Run(path+"/"+tc.name, func(t *testing.T) {
				repo := &adminRepository{}
				r := gin.New()
				NewHandlers(nil, repo, service.NewKeyManager()).RegisterRoutes(r, tc.configured)
				req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"name":"new","from":1,"to":2}`))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", tc.header)
				res := httptest.NewRecorder()
				r.ServeHTTP(res, req)
				if tc.allowed {
					if res.Code < 200 || res.Code >= 300 || repo.calls != 1 {
						t.Fatalf("authorized request: status=%d calls=%d body=%s", res.Code, repo.calls, res.Body)
					}
				} else if res.Code != http.StatusUnauthorized || repo.calls != 0 {
					t.Fatalf("unauthorized request: status=%d calls=%d", res.Code, repo.calls)
				}
			})
		}
	}
}

// TestPublicRoutesRemainReachable проверяет, что административный токен не требуется публичным маршрутам.
func TestPublicRoutesRemainReachable(t *testing.T) {
	r := gin.New()
	NewHandlers(nil, nil, nil).RegisterRoutes(r, "admin-secret")
	for _, tc := range []struct{ method, path string }{{"GET", "/service/invalid/key"}, {"POST", "/ticket"}} {
		res := httptest.NewRecorder()
		r.ServeHTTP(res, httptest.NewRequest(tc.method, tc.path, nil))
		if res.Code != http.StatusBadRequest {
			t.Fatalf("%s: status=%d", tc.path, res.Code)
		}
	}
}

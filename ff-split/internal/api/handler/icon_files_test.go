package handler

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/mock"
	iconservice "github.com/ivasnev/FinFlow/ff-split/internal/service/icon"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
	"github.com/stretchr/testify/require"
)

// iconTestValidator задаёт проверенные роли для транспортного теста.
type iconTestValidator struct{}

// ValidateToken отклоняет неизвестный токен и назначает тестовые роли.
func (iconTestValidator) ValidateToken(token string) (*auth.TokenPayload, error) {
	if token != "admin" && token != "user" {
		return nil, errors.New("invalid token")
	}
	return &auth.TokenPayload{UserID: 1, Roles: []string{token}}, nil
}

// TestIconFileGateway проверяет границы доступа и реальный multipart HTTP-вызов.
func TestIconFileGateway(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const fileID = "11111111-1111-4111-8111-111111111111"
	uploads := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Empty(t, r.Header.Get("Authorization"), "пользовательский JWT не должен передаваться файловому сервису")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			uploads++
			f, _, err := r.FormFile("file")
			require.NoError(t, err)
			defer f.Close()
			data, err := io.ReadAll(f)
			require.NoError(t, err)
			require.Equal(t, "image/png", http.DetectContentType(data))
			_, _ = io.WriteString(w, `{"data":{"object_id":"`+fileID+`"}}`)
		} else {
			_, _ = io.WriteString(w, `{"data":"https://storage.example/icon.png"}`)
		}
	}))
	defer upstream.Close()
	repo := mock.NewMockIcon(gomock.NewController(t))
	repo.EXPECT().GetIcons().Return([]models.Icon{{ID: 1, FileUUID: fileID}}, nil).AnyTimes()
	h := NewServerHandler(nil, nil, nil, nil, nil, nil, iconservice.NewIconService(repo), &IconFileGateway{Client: upstream.Client(), BaseURL: upstream.URL})
	router := gin.New()
	api.RegisterHandlersWithOptions(router, h, api.GinServerOptions{Middlewares: []api.MiddlewareFunc{api.MiddlewareFunc(auth.AuthMiddleware(iconTestValidator{}))}})
	for _, tc := range []struct {
		name, token string
		size, want  int
	}{
		{"без авторизации", "", 32, 401},
		{"обычный пользователь", "user", 32, 403},
		{"администратор", "admin", 32, 200},
		{"слишком большой файл", "admin", maxIconSize + 1, 413},
	} {
		// Сценарий: доступ зависит от роли, размер изображения ограничен до обращения к upstream.
		t.Run(tc.name, func(t *testing.T) {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			part, err := writer.CreateFormFile("file", "icon.png")
			require.NoError(t, err)
			data := make([]byte, tc.size)
			copy(data, []byte("\x89PNG\r\n\x1a\n"))
			_, err = part.Write(data)
			require.NoError(t, err)
			require.NoError(t, writer.Close())
			req := httptest.NewRequest(http.MethodPost, "/api/v1/icon-files", &body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			require.Equal(t, tc.want, response.Code, response.Body.String())
		})
	}
	require.Equal(t, 1, uploads)
	for _, operation := range []struct{ method, path string }{
		{http.MethodPost, "/api/v1/manage/icons"},
		{http.MethodPut, "/api/v1/manage/icons/1"},
		{http.MethodDelete, "/api/v1/manage/icons/1"},
	} {
		req := httptest.NewRequest(operation.method, operation.path, strings.NewReader(`{"name":"foreign","file_uuid":"22222222-2222-4222-8222-222222222222"}`))
		req.Header.Set("Authorization", "Bearer user")
		req.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusForbidden, response.Code, response.Body.String())
	}
	for _, tc := range []struct {
		path string
		want int
	}{
		{fileID, 200}, {"22222222-2222-4222-8222-222222222222", 404}, {"invalid", 400},
	} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/icon-files/"+tc.path, nil)
		req.Header.Set("Authorization", "Bearer user")
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		require.Equal(t, tc.want, response.Code, response.Body.String())
	}
}

// TestIconFileGatewayRejectsInvalidUpstream проверяет ограничение недоверенного ответа.
func TestIconFileGatewayRejectsInvalidUpstream(t *testing.T) {
	for _, body := range []string{"not json", strings.Repeat("x", (1<<20)+1)} {
		upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = io.WriteString(w, body) }))
		h := &ServerHandler{iconFiles: &IconFileGateway{Client: upstream.Client(), BaseURL: upstream.URL}}
		response := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(response)
		ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		h.forwardIconFile(ctx, http.MethodGet, "/api/v1/files/test", nil, "")
		require.Equal(t, http.StatusBadGateway, response.Code)
		upstream.Close()
	}
}

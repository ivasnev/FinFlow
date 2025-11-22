package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"
)

// LogoutSuite представляет suite для тестов выхода
type LogoutSuite struct {
	BaseSuite
}

// TestLogoutSuite запускает все тесты в LogoutSuite
func TestLogoutSuite(t *testing.T) {
	suite.Run(t, new(LogoutSuite))
}

// TestLogout_Success тестирует успешный выход
func (s *LogoutSuite) TestLogout_Success() {
	ctx := context.Background()

	// Сначала регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("logout@example.com"),
		Nickname: "logoutuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())
	s.NotNil(registerResp.JSON201)

	// Выполняем выход
	logoutReq := api.LogoutJSONRequestBody{
		RefreshToken: registerResp.JSON201.RefreshToken,
	}
	logoutResp, err := s.APIClient.LogoutWithResponse(ctx, logoutReq, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
		return nil
	})

	s.NoError(err, "выход должен пройти успешно")
	s.Equal(200, logoutResp.StatusCode(), "должен быть статус 200")

	// Проверяем, что токен больше не работает
	refreshReq := api.RefreshTokenJSONRequestBody{
		RefreshToken: registerResp.JSON201.RefreshToken,
	}
	refreshResp, err := s.APIClient.RefreshTokenWithResponse(ctx, refreshReq)
	s.NoError(err, "запрос должен выполниться успешно")
	s.Equal(401, refreshResp.StatusCode(), "должен быть статус 401")
	s.NotNil(refreshResp.JSON401, "должен быть возвращен объект ошибки")
	s.Contains(refreshResp.JSON401.Error, "недействительный refresh-токен")
}

// TestLogout_InvalidToken тестирует выход с недействительным токеном
func (s *LogoutSuite) TestLogout_InvalidToken() {
	ctx := context.Background()

	// Сначала регистрируем пользователя для получения валидного access токена
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("invalidtoken@example.com"),
		Nickname: "invalidtokenuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())
	s.NotNil(registerResp.JSON201)

	logoutReq := api.LogoutJSONRequestBody{
		RefreshToken: "invalid-refresh-token",
	}
	logoutResp, err := s.APIClient.LogoutWithResponse(ctx, logoutReq, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
		return nil
	})

	s.NoError(err, "запрос должен выполниться успешно")
	s.Equal(500, logoutResp.StatusCode(), "должен быть статус 500 (ошибка сервера при недействительном refresh токене)")
	s.NotNil(logoutResp.JSON500, "должен быть возвращен объект ошибки")
	s.Contains(logoutResp.JSON500.Error, "недействительный refresh-токен")
}

// TestLogout_AlreadyLoggedOut тестирует повторный выход
func (s *LogoutSuite) TestLogout_AlreadyLoggedOut() {
	ctx := context.Background()

	// Сначала регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("doublelogout@example.com"),
		Nickname: "doublelogoutuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())
	s.NotNil(registerResp.JSON201)

	// Первый выход
	logoutReq1 := api.LogoutJSONRequestBody{
		RefreshToken: registerResp.JSON201.RefreshToken,
	}
	logoutResp1, err := s.APIClient.LogoutWithResponse(ctx, logoutReq1, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
		return nil
	})
	s.NoError(err)
	s.Equal(200, logoutResp1.StatusCode())

	// Второй выход с тем же токеном должен вернуть ошибку
	// Используем access токен от первого выхода (он все еще валиден)
	logoutResp2, err := s.APIClient.LogoutWithResponse(ctx, logoutReq1, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
		return nil
	})
	s.NoError(err, "запрос должен выполниться успешно")
	s.Equal(500, logoutResp2.StatusCode(), "должен быть статус 500 (ошибка сервера при недействительном refresh токене)")
	s.NotNil(logoutResp2.JSON500, "должен быть возвращен объект ошибки")
	s.Contains(logoutResp2.JSON500.Error, "недействительный refresh-токен")
}


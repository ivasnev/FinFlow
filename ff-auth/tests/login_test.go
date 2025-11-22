package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"
)

// LoginSuite представляет suite для тестов входа
type LoginSuite struct {
	BaseSuite
}

// TestLoginSuite запускает все тесты в LoginSuite
func TestLoginSuite(t *testing.T) {
	suite.Run(t, new(LoginSuite))
}

// TestLogin_ByEmail тестирует успешный вход по email
func (s *LoginSuite) TestLogin_ByEmail() {
	ctx := context.Background()

	// Настройка мока для регистрации
	s.MockServer.
		Expect(http.MethodPost, "/api/v1/internal/users/register").
		Return("ff_id_service/register_user_response_success.json").
		HTTPCode(http.StatusCreated)

	// Сначала регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("login@example.com"),
		Nickname: "loginuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode(), "регистрация должна пройти успешно")

	// Выходим, чтобы можно было войти снова (регистрация создает сессию)
	if registerResp.JSON201 != nil {
		logoutReq := api.LogoutJSONRequestBody{
			RefreshToken: registerResp.JSON201.RefreshToken,
		}
		logoutResp, err := s.APIClient.LogoutWithResponse(ctx, logoutReq, func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
			return nil
		})
		if err == nil && logoutResp.StatusCode() == 200 {
			// Успешно вышли
		}
	}

	// Небольшая задержка для разного времени генерации токенов
	time.Sleep(100 * time.Millisecond)

	// Теперь выполняем вход
	reqBody := api.LoginJSONRequestBody{
		Login:    "login@example.com",
		Password: "password123",
	}

	resp, err := s.APIClient.LoginWithResponse(ctx, reqBody, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", "test-agent")
		return nil
	})
	s.NoError(err)

	if resp.StatusCode() != 200 {
		s.T().Logf("Ожидался статус 200, получен %d. Body: %s", resp.StatusCode(), string(resp.Body))
		if resp.JSON400 != nil {
			s.T().Logf("Ошибка 400: %s", resp.JSON400.Error)
		}
		if resp.JSON401 != nil {
			s.T().Logf("Ошибка 401: %s", resp.JSON401.Error)
		}
		if resp.JSON500 != nil {
			s.T().Logf("Ошибка 500: %s", resp.JSON500.Error)
		}
	}
	s.Equal(200, resp.StatusCode(), "вход должен пройти успешно")
	s.NotNil(resp.JSON200, "должен быть возвращен ответ")

	authResp := resp.JSON200
	s.NotEmpty(authResp.AccessToken, "должен быть возвращен access token")
	s.NotEmpty(authResp.RefreshToken, "должен быть возвращен refresh token")
	s.Equal("login@example.com", string(authResp.User.Email))
	s.Equal("loginuser", authResp.User.Nickname)
}

// TestLogin_ByNickname тестирует успешный вход по nickname
func (s *LoginSuite) TestLogin_ByNickname() {
	ctx := context.Background()

	// Настройка мока для регистрации
	s.MockServer.
		Expect(http.MethodPost, "/api/v1/internal/users/register").
		Return("ff_id_service/register_user_response_success.json").
		HTTPCode(http.StatusCreated)

	// Сначала регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("nickname@example.com"),
		Nickname: "nicknameuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())

	// Выходим, чтобы можно было войти снова
	if registerResp.JSON201 != nil {
		logoutReq := api.LogoutJSONRequestBody{
			RefreshToken: registerResp.JSON201.RefreshToken,
		}
		s.APIClient.LogoutWithResponse(ctx, logoutReq, func(ctx context.Context, req *http.Request) error {
			req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
			return nil
		})
		time.Sleep(100 * time.Millisecond)
	}

	// Теперь выполняем вход по nickname
	reqBody := api.LoginJSONRequestBody{
		Login:    "nicknameuser",
		Password: "password123",
	}

	resp, err := s.APIClient.LoginWithResponse(ctx, reqBody, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", "test-agent")
		return nil
	})
	s.NoError(err)
	s.Equal(200, resp.StatusCode(), "вход должен пройти успешно")
	s.NotNil(resp.JSON200, "должен быть возвращен ответ")

	authResp := resp.JSON200
	s.NotEmpty(authResp.AccessToken, "должен быть возвращен access token")
	s.Equal("nickname@example.com", string(authResp.User.Email))
	s.Equal("nicknameuser", authResp.User.Nickname)
}

// TestLogin_InvalidCredentials тестирует вход с неверными учетными данными
func (s *LoginSuite) TestLogin_InvalidCredentials() {
	ctx := context.Background()

	// Настройка мока для регистрации
	s.MockServer.
		Expect(http.MethodPost, "/api/v1/internal/users/register").
		Return("ff_id_service/register_user_response_success.json").
		HTTPCode(http.StatusCreated)

	// Сначала регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("invalid@example.com"),
		Nickname: "invaliduser",
		Password: "password123",
	}
	_, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)

	// Пытаемся войти с неверным паролем
	reqBody := api.LoginJSONRequestBody{
		Login:    "invalid@example.com",
		Password: "wrongpassword",
	}

	resp, err := s.APIClient.LoginWithResponse(ctx, reqBody, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", "test-agent")
		return nil
	})
	s.NoError(err)
	s.Equal(401, resp.StatusCode(), "должна быть ошибка 401")
	s.NotNil(resp.JSON401, "должен быть возвращен объект ошибки")
	s.Contains(resp.JSON401.Error, "неверный логин или пароль")
}

// TestLogin_UserNotFound тестирует вход несуществующего пользователя
func (s *LoginSuite) TestLogin_UserNotFound() {
	ctx := context.Background()

	reqBody := api.LoginJSONRequestBody{
		Login:    "nonexistent@example.com",
		Password: "password123",
	}

	resp, err := s.APIClient.LoginWithResponse(ctx, reqBody, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", "test-agent")
		return nil
	})
	s.NoError(err)
	s.Equal(401, resp.StatusCode(), "должна быть ошибка 401")
	s.NotNil(resp.JSON401, "должен быть возвращен объект ошибки")
	s.Contains(resp.JSON401.Error, "неверный логин или пароль")
}

// TestLogin_MultipleActiveSessions тестирует возможность наличия нескольких активных сессий
func (s *LoginSuite) TestLogin_MultipleActiveSessions() {
	ctx := context.Background()

	// Настройка мока для регистрации
	s.MockServer.
		Expect(http.MethodPost, "/api/v1/internal/users/register").
		Return("ff_id_service/register_user_response_success.json").
		HTTPCode(http.StatusCreated)

	// Регистрируем пользователя (создает первую сессию)
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("multisession@example.com"),
		Nickname: "multisessionuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode(), "регистрация должна пройти успешно")
	s.NotNil(registerResp.JSON201, "должен быть возвращен ответ")

	// Сохраняем данные первой сессии
	firstSessionRefreshToken := registerResp.JSON201.RefreshToken
	firstSessionAccessToken := registerResp.JSON201.AccessToken

	// Проверяем, что есть одна сессия после регистрации
	sessionsAfterRegisterResp, err := s.APIClient.GetUserSessionsWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+firstSessionAccessToken)
		return nil
	})
	s.NoError(err, "получение сессий должно пройти успешно")
	s.Equal(200, sessionsAfterRegisterResp.StatusCode())
	s.NotNil(sessionsAfterRegisterResp.JSON200)
	sessionsAfterRegister := *sessionsAfterRegisterResp.JSON200
	s.Len(sessionsAfterRegister, 1, "должна быть 1 сессия после регистрации")

	// Небольшая задержка для разного времени генерации токенов
	time.Sleep(100 * time.Millisecond)

	// Выполняем вход БЕЗ выхода из первой сессии (создает вторую сессию)
	loginReq := api.LoginJSONRequestBody{
		Login:    "multisession@example.com",
		Password: "password123",
	}
	loginResp, err := s.APIClient.LoginWithResponse(ctx, loginReq, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("User-Agent", "test-agent-2")
		return nil
	})
	s.NoError(err)
	s.Equal(200, loginResp.StatusCode(), "вход должен пройти успешно")
	s.NotNil(loginResp.JSON200, "должен быть возвращен ответ")

	// Сохраняем данные второй сессии
	secondSessionRefreshToken := loginResp.JSON200.RefreshToken
	secondSessionAccessToken := loginResp.JSON200.AccessToken

	// Проверяем, что refresh токены разные
	s.NotEqual(firstSessionRefreshToken, secondSessionRefreshToken, "refresh токены должны быть разными")
	s.NotEqual(firstSessionAccessToken, secondSessionAccessToken, "access токены должны быть разными")

	// Проверяем, что теперь у пользователя есть 2 активные сессии
	sessionsAfterLoginResp, err := s.APIClient.GetUserSessionsWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+secondSessionAccessToken)
		return nil
	})
	s.NoError(err, "получение сессий должно пройти успешно")
	s.Equal(200, sessionsAfterLoginResp.StatusCode())
	s.NotNil(sessionsAfterLoginResp.JSON200)
	sessionsAfterLogin := *sessionsAfterLoginResp.JSON200
	s.Len(sessionsAfterLogin, 2, "должно быть 2 активные сессии")

	// Проверяем, что обе сессии имеют разные ID
	s.NotEqual(sessionsAfterLogin[0].Id, sessionsAfterLogin[1].Id, "сессии должны иметь разные ID")

	// Проверяем, что обе сессии принадлежат одному пользователю и активны
	for _, session := range sessionsAfterLogin {
		s.True(session.ExpiresAt.After(time.Now()), "сессия должна быть активной (не истекшей)")
	}

	// Проверяем, что обе сессии можно использовать для refresh токена
	// Используем первую сессию
	refreshResp1, err := s.APIClient.RefreshTokenWithResponse(ctx, api.RefreshTokenJSONRequestBody{
		RefreshToken: firstSessionRefreshToken,
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+firstSessionAccessToken)
		return nil
	})
	s.NoError(err)
	s.Equal(200, refreshResp1.StatusCode(), "refresh первой сессии должен пройти успешно")

	// Используем вторую сессию
	refreshResp2, err := s.APIClient.RefreshTokenWithResponse(ctx, api.RefreshTokenJSONRequestBody{
		RefreshToken: secondSessionRefreshToken,
	}, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+secondSessionAccessToken)
		return nil
	})
	s.NoError(err)
	s.Equal(200, refreshResp2.StatusCode(), "refresh второй сессии должен пройти успешно")

	// Проверяем, что после refresh все еще есть 2 сессии (старые удаляются, новые создаются)
	sessionsAfterRefreshResp, err := s.APIClient.GetUserSessionsWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+refreshResp2.JSON200.AccessToken)
		return nil
	})
	s.NoError(err, "получение сессий должно пройти успешно")
	s.Equal(200, sessionsAfterRefreshResp.StatusCode())
	s.NotNil(sessionsAfterRefreshResp.JSON200)
	sessionsAfterRefresh := *sessionsAfterRefreshResp.JSON200
	s.Len(sessionsAfterRefresh, 2, "должно остаться 2 активные сессии после refresh")
}

package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"
)

// SessionsSuite представляет suite для тестов управления сессиями
type SessionsSuite struct {
	BaseSuite
}

// TestSessionsSuite запускает все тесты в SessionsSuite
func TestSessionsSuite(t *testing.T) {
	suite.Run(t, new(SessionsSuite))
}

// TestGetUserSessions_Success тестирует успешное получение сессий пользователя
func (s *SessionsSuite) TestGetUserSessions_Success() {
	ctx := context.Background()

	// Регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("sessions@example.com"),
		Nickname: "sessionsuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())
	s.NotNil(registerResp.JSON201)

	// Получаем все сессии пользователя (должна быть минимум одна от регистрации)
	sessionsResp, err := s.APIClient.GetUserSessionsWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
		return nil
	})

	s.NoError(err, "получение сессий должно пройти успешно")
	s.Equal(200, sessionsResp.StatusCode(), "должен быть статус 200")
	s.NotNil(sessionsResp.JSON200, "сессии должны быть возвращены")
	s.GreaterOrEqual(len(*sessionsResp.JSON200), 1, "должна быть минимум 1 сессия")
}

// TestGetUserSessions_Empty тестирует получение сессий для пользователя без сессий
// Примечание: API не предоставляет метод для завершения всех сессий, поэтому используем сервис
func (s *SessionsSuite) TestGetUserSessions_Empty() {
	ctx := context.Background()

	// Регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("nosessions@example.com"),
		Nickname: "nosessionsuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())
	s.NotNil(registerResp.JSON201)

	// Удаляем все сессии через сервис (API не предоставляет этот метод)
	err = s.Container.SessionService.TerminateAllSessions(ctx, registerResp.JSON201.User.Id)
	s.NoError(err)

	// Получаем сессии через API
	sessionsResp, err := s.APIClient.GetUserSessionsWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
		return nil
	})

	s.NoError(err, "получение сессий должно пройти успешно")
	s.Equal(200, sessionsResp.StatusCode(), "должен быть статус 200")
	s.NotNil(sessionsResp.JSON200, "сессии должны быть возвращены")
	s.Empty(*sessionsResp.JSON200, "сессий не должно быть")
}

// TestTerminateSession_Success тестирует успешное завершение сессии
func (s *SessionsSuite) TestTerminateSession_Success() {
	ctx := context.Background()

	// Регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("terminate@example.com"),
		Nickname: "terminateuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())
	s.NotNil(registerResp.JSON201)

	// Получаем сессии
	sessionsResp, err := s.APIClient.GetUserSessionsWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
		return nil
	})
	s.NoError(err)
	s.Equal(200, sessionsResp.StatusCode())
	s.NotNil(sessionsResp.JSON200)
	sessions := *sessionsResp.JSON200
	s.NotEmpty(sessions, "должна быть хотя бы одна сессия")

	// Завершаем первую сессию
	terminateResp, err := s.APIClient.TerminateSessionWithResponse(ctx, sessions[0].Id, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
		return nil
	})
	s.NoError(err, "завершение сессии должно пройти успешно")
	s.Equal(200, terminateResp.StatusCode(), "должен быть статус 200")

	// Проверяем, что сессия удалена
	sessionsAfterResp, err := s.APIClient.GetUserSessionsWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
		return nil
	})
	s.NoError(err)
	s.Equal(200, sessionsAfterResp.StatusCode())
	s.NotNil(sessionsAfterResp.JSON200)
	sessionsAfter := *sessionsAfterResp.JSON200
	s.Len(sessionsAfter, len(sessions)-1, "количество сессий должно уменьшиться")
}

// TestTerminateAllSessions_Success тестирует успешное завершение всех сессий
// Примечание: API не предоставляет метод для завершения всех сессий, поэтому используем сервис
func (s *SessionsSuite) TestTerminateAllSessions_Success() {
	ctx := context.Background()

	// Регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("terminateall@example.com"),
		Nickname: "terminatealluser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())
	s.NotNil(registerResp.JSON201)

	// Проверяем, что есть сессии через API (минимум одна от регистрации)
	sessionsBeforeResp, err := s.APIClient.GetUserSessionsWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
		return nil
	})
	s.NoError(err)
	s.Equal(200, sessionsBeforeResp.StatusCode())
	s.NotNil(sessionsBeforeResp.JSON200)
	sessionsBefore := *sessionsBeforeResp.JSON200
	s.GreaterOrEqual(len(sessionsBefore), 1, "должна быть минимум 1 сессия")

	// Завершаем все сессии через сервис (API не предоставляет этот метод)
	err = s.Container.SessionService.TerminateAllSessions(ctx, registerResp.JSON201.User.Id)
	s.NoError(err, "завершение всех сессий должно пройти успешно")

	// Проверяем, что все сессии удалены через API
	sessionsAfterResp, err := s.APIClient.GetUserSessionsWithResponse(ctx, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
		return nil
	})
	s.NoError(err)
	s.Equal(200, sessionsAfterResp.StatusCode())
	s.NotNil(sessionsAfterResp.JSON200)
	sessionsAfter := *sessionsAfterResp.JSON200
	s.Empty(sessionsAfter, "все сессии должны быть удалены")
}

package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	"github.com/ivasnev/FinFlow/ff-auth/tests/mockserver"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"
)

// RegisterSuite представляет suite для тестов регистрации
type RegisterSuite struct {
	BaseSuite
}

// TestRegisterSuite запускает все тесты в RegisterSuite
func TestRegisterSuite(t *testing.T) {
	suite.Run(t, new(RegisterSuite))
}

// TestRegister_Success тестирует успешную регистрацию пользователя
func (s *RegisterSuite) TestRegister_Success() {
	ctx := context.Background()

	reqBody := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("test@example.com"),
		Nickname: "testuser",
		Password: "password123",
		Name:     stringPtr("Test User"),
	}

	resp, err := s.APIClient.RegisterWithResponse(ctx, reqBody)
	s.NoError(err)
	s.Equal(201, resp.StatusCode(), "должен быть статус 201")
	s.NotNil(resp.JSON201, "должен быть возвращен ответ")

	authResp := resp.JSON201
	s.NotEmpty(authResp.AccessToken, "должен быть возвращен access token")
	s.NotEmpty(authResp.RefreshToken, "должен быть возвращен refresh token")
	s.Equal("test@example.com", string(authResp.User.Email))
	s.Equal("testuser", authResp.User.Nickname)
	s.Greater(authResp.User.Id, int64(0), "ID пользователя должен быть больше 0")

	// Проверяем, что пользователь создан в БД
	var count int64
	err = s.GetDB().Table("users").Where("email = ?", "test@example.com").Count(&count).Error
	s.NoError(err, "пользователь должен быть создан в БД")
	s.Equal(int64(1), count, "должен быть создан один пользователь")
}

// TestRegister_DuplicateEmail тестирует регистрацию с дублирующимся email
func (s *RegisterSuite) TestRegister_DuplicateEmail() {
	ctx := context.Background()

	reqBody := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("duplicate@example.com"),
		Nickname: "user1",
		Password: "password123",
	}

	// Первая регистрация
	resp1, err := s.APIClient.RegisterWithResponse(ctx, reqBody)
	s.NoError(err)
	s.Equal(201, resp1.StatusCode())

	// Вторая регистрация с тем же email
	reqBody2 := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("duplicate@example.com"),
		Nickname: "user2",
		Password: "password123",
	}
	resp2, err := s.APIClient.RegisterWithResponse(ctx, reqBody2)
	s.NoError(err)
	s.Equal(500, resp2.StatusCode(), "должна быть ошибка 500")
	s.NotNil(resp2.JSON500, "должен быть возвращен объект ошибки")
	s.Contains(resp2.JSON500.Error, "уже существует")
}

// TestRegister_DuplicateNickname тестирует регистрацию с дублирующимся nickname
func (s *RegisterSuite) TestRegister_DuplicateNickname() {
	ctx := context.Background()

	reqBody := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("user1@example.com"),
		Nickname: "duplicatenick",
		Password: "password123",
	}

	// Первая регистрация
	resp1, err := s.APIClient.RegisterWithResponse(ctx, reqBody)
	s.NoError(err)
	s.Equal(201, resp1.StatusCode())

	// Вторая регистрация с тем же nickname
	reqBody2 := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("user2@example.com"),
		Nickname: "duplicatenick",
		Password: "password123",
	}
	resp2, err := s.APIClient.RegisterWithResponse(ctx, reqBody2)
	s.NoError(err)
	s.Equal(500, resp2.StatusCode(), "должна быть ошибка 500")
	s.NotNil(resp2.JSON500, "должен быть возвращен объект ошибки")
	s.Contains(resp2.JSON500.Error, "уже существует")
}

// TestRegister_WithMockServerError тестирует обработку ошибок от мок-сервера
func (s *RegisterSuite) TestRegister_WithMockServerError() {
	// Устанавливаем обработчик, который возвращает ошибку 500
	errorHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(mockserver.ErrorResponse{
			Error: "внутренняя ошибка сервера",
		})
	}
	s.MockServer.SetRegisterUserFromServiceHandler(errorHandler)

	ctx := context.Background()

	reqBody := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("test@example.com"),
		Nickname: "testuser",
		Password: "password123",
	}

	resp, err := s.APIClient.RegisterWithResponse(ctx, reqBody)
	s.NoError(err)
	s.Equal(500, resp.StatusCode(), "должна быть ошибка 500")
	s.NotNil(resp.JSON500, "должен быть возвращен объект ошибки")
	s.Contains(resp.JSON500.Error, "ID", "ошибка должна быть связана с ID сервисом")

	// Проверяем, что пользователь не создан в БД (должен быть откат транзакции)
	var count int64
	err = s.GetDB().Table("users").Where("email = ?", "test@example.com").Count(&count).Error
	s.NoError(err)
	s.Equal(int64(0), count, "пользователь не должен быть создан в БД")
}

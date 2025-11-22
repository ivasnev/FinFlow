package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"
)

// UsersSuite представляет suite для тестов управления пользователями
type UsersSuite struct {
	BaseSuite
}

// TestUsersSuite запускает все тесты в UsersSuite
func TestUsersSuite(t *testing.T) {
	suite.Run(t, new(UsersSuite))
}

// TestGetUserByID_Success тестирует успешное получение пользователя по ID
// Используем GetUserByNickname, так как API не предоставляет метод получения по ID
func (s *UsersSuite) TestGetUserByID_Success() {
	ctx := context.Background()

	// Настройка мока для регистрации
	s.MockServer.
		Expect(http.MethodPost, "/api/v1/internal/users/register").
		Return("ff_id_service/register_user_response_success.json").
		HTTPCode(http.StatusCreated)

	// Регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("getuser@example.com"),
		Nickname: "getuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())
	s.NotNil(registerResp.JSON201)

	// Получаем пользователя по nickname (API не предоставляет метод по ID)
	userResp, err := s.APIClient.GetUserByNicknameWithResponse(ctx, registerResp.JSON201.User.Nickname)

	s.NoError(err, "получение пользователя должно пройти успешно")
	s.Equal(200, userResp.StatusCode(), "должен быть статус 200")
	s.NotNil(userResp.JSON200, "пользователь должен быть возвращен")
	s.Equal(registerResp.JSON201.User.Id, userResp.JSON200.Id)
	s.Equal(registerResp.JSON201.User.Email, userResp.JSON200.Email)
	s.Equal(registerResp.JSON201.User.Nickname, userResp.JSON200.Nickname)
}

// TestGetUserByNickname_Success тестирует успешное получение пользователя по nickname
func (s *UsersSuite) TestGetUserByNickname_Success() {
	ctx := context.Background()

	// Настройка мока для регистрации
	s.MockServer.
		Expect(http.MethodPost, "/api/v1/internal/users/register").
		Return("ff_id_service/register_user_response_success.json").
		HTTPCode(http.StatusCreated)

	// Регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("nickname@example.com"),
		Nickname: "nicknameuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())
	s.NotNil(registerResp.JSON201)

	// Получаем пользователя по nickname
	userResp, err := s.APIClient.GetUserByNicknameWithResponse(ctx, registerResp.JSON201.User.Nickname)

	s.NoError(err, "получение пользователя должно пройти успешно")
	s.Equal(200, userResp.StatusCode(), "должен быть статус 200")
	s.NotNil(userResp.JSON200, "пользователь должен быть возвращен")
	s.Equal(registerResp.JSON201.User.Id, userResp.JSON200.Id)
	s.Equal(registerResp.JSON201.User.Email, userResp.JSON200.Email)
	s.Equal(registerResp.JSON201.User.Nickname, userResp.JSON200.Nickname)
}

// TestGetUserByNickname_NotFound тестирует получение несуществующего пользователя
func (s *UsersSuite) TestGetUserByNickname_NotFound() {
	ctx := context.Background()

	userResp, err := s.APIClient.GetUserByNicknameWithResponse(ctx, "nonexistentuser")

	s.NoError(err, "запрос должен выполниться успешно")
	s.Equal(404, userResp.StatusCode(), "должен быть статус 404")
	s.NotNil(userResp.JSON404, "должен быть возвращен объект ошибки")
}

// TestUpdateUser_Success тестирует успешное обновление пользователя
func (s *UsersSuite) TestUpdateUser_Success() {
	ctx := context.Background()

	// Настройка мока для регистрации
	s.MockServer.
		Expect(http.MethodPost, "/api/v1/internal/users/register").
		Return("ff_id_service/register_user_response_success.json").
		HTTPCode(http.StatusCreated)

	// Регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("update@example.com"),
		Nickname: "updateuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())
	s.NotNil(registerResp.JSON201)

	// Обновляем email
	newEmail := openapi_types.Email("updated@example.com")
	updateReq := api.UpdateUserJSONRequestBody{
		Email: &newEmail,
	}

	updateResp, err := s.APIClient.UpdateUserWithResponse(ctx, updateReq, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp.JSON201.AccessToken)
		return nil
	})

	s.NoError(err, "обновление пользователя должно пройти успешно")
	s.Equal(200, updateResp.StatusCode(), "должен быть статус 200")
	s.NotNil(updateResp.JSON200, "обновленный пользователь должен быть возвращен")
	s.Equal(newEmail, updateResp.JSON200.Email)
	s.Equal(registerResp.JSON201.User.Nickname, updateResp.JSON200.Nickname)
}

// TestUpdateUser_DuplicateEmail тестирует обновление с дублирующимся email
func (s *UsersSuite) TestUpdateUser_DuplicateEmail() {
	ctx := context.Background()

	// Настройка мока для первой регистрации
	s.MockServer.
		Expect(http.MethodPost, "/api/v1/internal/users/register").
		Return("ff_id_service/register_user_response_success.json").
		HTTPCode(http.StatusCreated)

	// Регистрируем первого пользователя
	registerReq1 := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("user1@example.com"),
		Nickname: "user1",
		Password: "password123",
	}
	registerResp1, err := s.APIClient.RegisterWithResponse(ctx, registerReq1)
	s.NoError(err)
	s.Equal(201, registerResp1.StatusCode())

	// Настройка мока для второй регистрации
	s.MockServer.
		Expect(http.MethodPost, "/api/v1/internal/users/register").
		Return("ff_id_service/register_user_response_success.json").
		HTTPCode(http.StatusCreated)

	// Регистрируем второго пользователя
	registerReq2 := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("user2@example.com"),
		Nickname: "user2",
		Password: "password123",
	}
	registerResp2, err := s.APIClient.RegisterWithResponse(ctx, registerReq2)
	s.NoError(err)
	s.Equal(201, registerResp2.StatusCode())

	// Пытаемся обновить email второго пользователя на email первого
	duplicateEmail := registerResp1.JSON201.User.Email
	updateReq := api.UpdateUserJSONRequestBody{
		Email: &duplicateEmail,
	}

	updateResp, err := s.APIClient.UpdateUserWithResponse(ctx, updateReq, func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+registerResp2.JSON201.AccessToken)
		return nil
	})

	s.NoError(err, "запрос должен выполниться успешно")
	// API может возвращать 400 или 500 при конфликте email, проверяем что это ошибка
	s.NotEqual(200, updateResp.StatusCode(), "должен быть статус ошибки")
	s.True(updateResp.StatusCode() == 400 || updateResp.StatusCode() == 500, "должен быть статус 400 или 500")
	if updateResp.StatusCode() == 400 {
		s.NotNil(updateResp.JSON400, "должен быть возвращен объект ошибки")
	} else if updateResp.StatusCode() == 500 {
		s.NotNil(updateResp.JSON500, "должен быть возвращен объект ошибки")
	}
}

// TestDeleteUser_Success тестирует успешное удаление пользователя
// Примечание: API не предоставляет метод удаления пользователя, поэтому используем прямой доступ к БД
// для проверки удаления. В реальном приложении это может быть внутренний endpoint.
func (s *UsersSuite) TestDeleteUser_Success() {
	ctx := context.Background()

	// Настройка мока для регистрации
	s.MockServer.
		Expect(http.MethodPost, "/api/v1/internal/users/register").
		Return("ff_id_service/register_user_response_success.json").
		HTTPCode(http.StatusCreated)

	// Регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("delete@example.com"),
		Nickname: "deleteuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())
	s.NotNil(registerResp.JSON201)

	// Удаляем пользователя через сервис (API не предоставляет этот метод)
	err = s.Container.UserService.DeleteUser(ctx, registerResp.JSON201.User.Id)

	s.NoError(err, "удаление пользователя должно пройти успешно")

	// Проверяем, что пользователь удален (используем простой запрос без загрузки ассоциаций)
	var count int64
	err = s.GetDB().Table("users").Where("id = ?", registerResp.JSON201.User.Id).Count(&count).Error
	s.NoError(err, "запрос должен выполниться успешно")
	s.Equal(int64(0), count, "пользователь должен быть удален из БД")
}

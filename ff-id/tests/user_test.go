package tests

import (
	"testing"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
	"github.com/ivasnev/FinFlow/ff-id/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"
)

// UserSuite представляет suite для тестов управления пользователями
type UserSuite struct {
	BaseSuite
}

// TestUserSuite запускает все тесты в UserSuite
func TestUserSuite(t *testing.T) {
	suite.Run(t, new(UserSuite))
}

// createTestUser создает тестового пользователя в БД с использованием константного времени
func (s *UserSuite) createTestUser(id int64, email, nickname, name string) *models.User {
	now := s.TimeProvider.Now()

	user := &models.User{
		ID:        id,
		Email:     email,
		Nickname:  nickname,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if name != "" {
		user.Name.String = name
		user.Name.Valid = true
	}

	// Создаем пользователя напрямую в БД
	err := s.GetDB().Exec(`
		INSERT INTO users (id, email, nickname, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.Email, user.Nickname, user.Name, user.CreatedAt, user.UpdatedAt).Error
	s.NoError(err, "не удалось создать тестового пользователя")

	return user
}

// TestGetUserByNickname_Success тестирует успешное получение пользователя по nickname
func (s *UserSuite) TestGetUserByNickname_Success() {
	// Arrange - подготовка
	testUser := s.createTestUser(TestUserID1, TestEmail1, TestNickname1, TestName1)

	// Act - действие
	resp, err := s.APIClient.GetUserByNicknameWithResponse(s.Ctx, testUser.Nickname)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "пользователь должен быть возвращен")
	s.Require().Equal(testUser.ID, resp.JSON200.Id)
	s.Require().Equal(testUser.Email, string(resp.JSON200.Email))
	s.Require().Equal(testUser.Nickname, resp.JSON200.Nickname)
}

// TestGetUserByNickname_NotFound тестирует получение несуществующего пользователя
func (s *UserSuite) TestGetUserByNickname_NotFound() {
	// Arrange - подготовка
	nonExistentNickname := "nonexistentuser"

	// Act - действие
	resp, err := s.APIClient.GetUserByNicknameWithResponse(s.Ctx, nonExistentNickname)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(404, resp.StatusCode(), "должен быть статус 404")
	s.Require().NotNil(resp.JSON404, "должен быть возвращен объект ошибки")
}

// TestRegisterUser_Success тестирует успешную регистрацию пользователя
func (s *UserSuite) TestRegisterUser_Success() {
	// Используем s.Ctx из BaseSuite

	name := "New User"
	phone := "+79123456789"
	reqBody := api.RegisterUserJSONRequestBody{
		Email:    openapi_types.Email("newuser@example.com"),
		Nickname: "newuser",
		Name:     &name,
		Phone:    &phone,
	}

	// Для этого теста нужно будет добавить mock для auth middleware
	// или вызвать метод сервиса напрямую
	var avatarID *uuid.UUID
	if reqBody.AvatarId != nil {
		id := uuid.UUID(*reqBody.AvatarId)
		avatarID = &id
	}

	userDTO, err := s.Container.UserService.RegisterUser(s.Ctx, 100, &service.RegisterUserRequest{
		Email:     string(reqBody.Email),
		Nickname:  reqBody.Nickname,
		Name:      reqBody.Name,
		Phone:     reqBody.Phone,
		Birthdate: reqBody.Birthdate,
		AvatarID:  avatarID,
	})

	s.NoError(err, "регистрация должна пройти успешно")
	s.NotNil(userDTO, "должен быть возвращен пользователь")
	s.Equal(int64(100), userDTO.ID)
	s.Equal("newuser@example.com", userDTO.Email)
	s.Equal("newuser", userDTO.Nickname)

	// Проверяем, что пользователь создан в БД
	var count int64
	err = s.GetDB().Table("users").Where("email = ?", "newuser@example.com").Count(&count).Error
	s.NoError(err, "пользователь должен быть создан в БД")
	s.Equal(int64(1), count, "должен быть создан один пользователь")
}

// TestRegisterUser_DuplicateEmail тестирует регистрацию с дублирующимся email
func (s *UserSuite) TestRegisterUser_DuplicateEmail() {
	// Используем s.Ctx из BaseSuite

	// Создаем первого пользователя
	s.createTestUser(1, "duplicate@example.com", "user1", "User One")

	// Пытаемся создать второго пользователя с тем же email
	_, err := s.Container.UserService.RegisterUser(s.Ctx, 2, &service.RegisterUserRequest{
		Email:    "duplicate@example.com",
		Nickname: "user2",
	})

	s.Error(err, "должна быть ошибка при дублировании email")
	// Проверяем, что ошибка содержит информацию о дублировании
	errMsg := err.Error()
	s.True(
		s.Contains(errMsg, "уже используется") || s.Contains(errMsg, "уже существует"),
		"ошибка должна содержать информацию о дублировании, получено: %s", errMsg,
	)
}

// TestRegisterUser_DuplicateNickname тестирует регистрацию с дублирующимся nickname
func (s *UserSuite) TestRegisterUser_DuplicateNickname() {
	// Используем s.Ctx из BaseSuite

	// Создаем первого пользователя
	s.createTestUser(1, "user1@example.com", "duplicatenick", "User One")

	// Пытаемся создать второго пользователя с тем же nickname
	_, err := s.Container.UserService.RegisterUser(s.Ctx, 2, &service.RegisterUserRequest{
		Email:    "user2@example.com",
		Nickname: "duplicatenick",
	})

	s.Error(err, "должна быть ошибка при дублировании nickname")
	// Проверяем, что ошибка содержит информацию о дублировании
	errMsg := err.Error()
	s.True(
		s.Contains(errMsg, "уже используется") || s.Contains(errMsg, "уже существует"),
		"ошибка должна содержать информацию о дублировании, получено: %s", errMsg,
	)
}

// TestUpdateUser_Success тестирует успешное обновление пользователя
func (s *UserSuite) TestUpdateUser_Success() {
	// Используем s.Ctx из BaseSuite

	// Создаем тестового пользователя
	testUser := s.createTestUser(1, "update@example.com", "updateuser", "Old Name")

	// Обновляем данные пользователя
	newEmail := "updated@example.com"
	newName := "New Name"
	newPhone := "+79876543210"

	userDTO, err := s.Container.UserService.UpdateUser(s.Ctx, testUser.ID, service.UpdateUserRequest{
		Email: &newEmail,
		Name:  &newName,
		Phone: &newPhone,
	})

	s.NoError(err, "обновление должно пройти успешно")
	s.NotNil(userDTO, "должен быть возвращен пользователь")
	s.Equal(newEmail, userDTO.Email)
	s.Equal(newName, *userDTO.Name)
	s.Equal(newPhone, *userDTO.Phone)
	s.Equal(testUser.Nickname, userDTO.Nickname)

	// Проверяем, что данные обновлены в БД
	var updatedUser models.User
	err = s.GetDB().Table("users").Where("id = ?", testUser.ID).First(&updatedUser).Error
	s.NoError(err, "пользователь должен быть в БД")
	s.Equal(newEmail, updatedUser.Email)
	s.Equal(newName, updatedUser.Name.String)
	s.Equal(newPhone, updatedUser.Phone.String)
}

// TestUpdateUser_DuplicateEmail тестирует обновление с дублирующимся email
func (s *UserSuite) TestUpdateUser_DuplicateEmail() {
	// Используем s.Ctx из BaseSuite

	// Создаем двух пользователей
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")
	user2 := s.createTestUser(2, "user2@example.com", "user2", "User Two")

	// Пытаемся обновить email второго пользователя на email первого
	_, err := s.Container.UserService.UpdateUser(s.Ctx, user2.ID, service.UpdateUserRequest{
		Email: &user1.Email,
	})

	s.Error(err, "должна быть ошибка при дублировании email")
	// Проверяем, что ошибка содержит информацию о дублировании
	errMsg := err.Error()
	s.True(
		s.Contains(errMsg, "уже используется") || s.Contains(errMsg, "уже существует"),
		"ошибка должна содержать информацию о дублировании, получено: %s", errMsg,
	)
}

// TestUpdateUser_DuplicateNickname тестирует обновление с дублирующимся nickname
func (s *UserSuite) TestUpdateUser_DuplicateNickname() {
	// Используем s.Ctx из BaseSuite

	// Создаем двух пользователей
	user1 := s.createTestUser(1, "user1@example.com", "nick1", "User One")
	user2 := s.createTestUser(2, "user2@example.com", "nick2", "User Two")

	// Пытаемся обновить nickname второго пользователя на nickname первого
	_, err := s.Container.UserService.UpdateUser(s.Ctx, user2.ID, service.UpdateUserRequest{
		Nickname: &user1.Nickname,
	})

	s.Error(err, "должна быть ошибка при дублировании nickname")
	// Проверяем, что ошибка содержит информацию о дублировании
	errMsg := err.Error()
	s.True(
		s.Contains(errMsg, "уже используется") || s.Contains(errMsg, "уже существует"),
		"ошибка должна содержать информацию о дублировании, получено: %s", errMsg,
	)
}

// TestGetUsersByIds_Success тестирует успешное получение пользователей по ID
func (s *UserSuite) TestGetUsersByIds_Success() {
	// Используем s.Ctx из BaseSuite

	// Создаем тестовых пользователей
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")
	user2 := s.createTestUser(2, "user2@example.com", "user2", "User Two")
	user3 := s.createTestUser(3, "user3@example.com", "user3", "User Three")

	// Получаем пользователей по ID
	users, err := s.Container.UserService.GetUsersByIds(s.Ctx, []int64{user1.ID, user2.ID, user3.ID})

	s.NoError(err, "получение пользователей должно пройти успешно")
	s.Len(users, 3, "должно быть возвращено 3 пользователя")

	// Проверяем, что все пользователи присутствуют
	userIDs := make(map[int64]bool)
	for _, user := range users {
		userIDs[user.ID] = true
	}
	s.True(userIDs[user1.ID], "user1 должен быть в результатах")
	s.True(userIDs[user2.ID], "user2 должен быть в результатах")
	s.True(userIDs[user3.ID], "user3 должен быть в результатах")
}

// TestGetUsersByIds_PartialResults тестирует получение пользователей когда некоторые ID не существуют
func (s *UserSuite) TestGetUsersByIds_PartialResults() {
	// Используем s.Ctx из BaseSuite

	// Создаем только одного пользователя
	user1 := s.createTestUser(1, "user1@example.com", "user1", "User One")

	// Запрашиваем пользователей, включая несуществующие ID
	users, err := s.Container.UserService.GetUsersByIds(s.Ctx, []int64{user1.ID, 999, 1000})

	s.NoError(err, "получение пользователей должно пройти успешно")
	s.Len(users, 1, "должен быть возвращен только 1 существующий пользователь")
	s.Equal(user1.ID, users[0].ID)
}

// TestGetUsersByIds_EmptyList тестирует получение пользователей с пустым списком ID
func (s *UserSuite) TestGetUsersByIds_EmptyList() {
	// Используем s.Ctx из BaseSuite

	users, err := s.Container.UserService.GetUsersByIds(s.Ctx, []int64{})

	s.NoError(err, "получение пользователей должно пройти успешно")
	s.Len(users, 0, "должен быть возвращен пустой список")
}

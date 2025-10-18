package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/mock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
)

func TestUserService_GetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockUser(ctrl)
	userService := NewUserService(mockRepo)

	ctx := context.Background()
	userID := int64(1)

	t.Run("успешное получение пользователя", func(t *testing.T) {
		user := &models.User{
			ID:        userID,
			Email:     "test@example.com",
			Nickname:  "testuser",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}

		roles := []models.RoleEntity{
			{ID: 1, Name: "user"},
			{ID: 2, Name: "admin"},
		}

		mockRepo.EXPECT().
			GetByID(ctx, userID).
			Return(user, nil).
			Times(1)

		mockRepo.EXPECT().
			GetRoles(ctx, userID).
			Return(roles, nil).
			Times(1)

		result, err := userService.GetUserByID(ctx, userID)

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if result == nil {
			t.Fatal("Ожидались данные пользователя, получен nil")
		}

		if result.Id != userID {
			t.Errorf("Ожидался ID %d, получен %d", userID, result.Id)
		}

		if result.Email != "test@example.com" {
			t.Errorf("Ожидался email 'test@example.com', получен '%s'", result.Email)
		}

		if result.Nickname != "testuser" {
			t.Errorf("Ожидался nickname 'testuser', получен '%s'", result.Nickname)
		}

		if len(result.Roles) != 2 {
			t.Errorf("Ожидалось 2 роли, получено %d", len(result.Roles))
		}

		expectedRoles := []string{"user", "admin"}
		for i, role := range result.Roles {
			if role != expectedRoles[i] {
				t.Errorf("Ожидалась роль '%s', получена '%s'", expectedRoles[i], role)
			}
		}
	})

	t.Run("ошибка получения пользователя", func(t *testing.T) {
		expectedErr := errors.New("user not found")
		mockRepo.EXPECT().
			GetByID(ctx, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := userService.GetUserByID(ctx, userID)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if result != nil {
			t.Fatal("Ожидался nil результат при ошибке")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
	})

	t.Run("ошибка получения ролей", func(t *testing.T) {
		user := &models.User{
			ID:       userID,
			Email:    "test@example.com",
			Nickname: "testuser",
		}

		mockRepo.EXPECT().
			GetByID(ctx, userID).
			Return(user, nil).
			Times(1)

		expectedErr := errors.New("roles not found")
		mockRepo.EXPECT().
			GetRoles(ctx, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := userService.GetUserByID(ctx, userID)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if result != nil {
			t.Fatal("Ожидался nil результат при ошибке")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
	})
}

func TestUserService_GetUserByNickname(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockUser(ctrl)
	userService := NewUserService(mockRepo)

	ctx := context.Background()
	nickname := "testuser"

	t.Run("успешное получение пользователя по никнейму", func(t *testing.T) {
		user := &models.User{
			ID:        1,
			Email:     "test@example.com",
			Nickname:  nickname,
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}

		roles := []models.RoleEntity{
			{ID: 1, Name: "user"},
		}

		mockRepo.EXPECT().
			GetByNickname(ctx, nickname).
			Return(user, nil).
			Times(1)

		mockRepo.EXPECT().
			GetRoles(ctx, user.ID).
			Return(roles, nil).
			Times(1)

		result, err := userService.GetUserByNickname(ctx, nickname)

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if result == nil {
			t.Fatal("Ожидались данные пользователя, получен nil")
		}

		if result.Nickname != nickname {
			t.Errorf("Ожидался nickname '%s', получен '%s'", nickname, result.Nickname)
		}
	})

	t.Run("ошибка получения пользователя по никнейму", func(t *testing.T) {
		expectedErr := errors.New("user not found")
		mockRepo.EXPECT().
			GetByNickname(ctx, nickname).
			Return(nil, expectedErr).
			Times(1)

		result, err := userService.GetUserByNickname(ctx, nickname)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if result != nil {
			t.Fatal("Ожидался nil результат при ошибке")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockUser(ctrl)
	userService := NewUserService(mockRepo)

	ctx := context.Background()
	userID := int64(1)

	t.Run("успешное обновление email", func(t *testing.T) {
		oldUser := &models.User{
			ID:        userID,
			Email:     "old@example.com",
			Nickname:  "testuser",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}

		newEmail := "new@example.com"
		updateData := service.UserUpdateData{
			Email: &newEmail,
		}

		roles := []models.RoleEntity{
			{ID: 1, Name: "user"},
		}

		// Получение пользователя
		mockRepo.EXPECT().
			GetByID(ctx, userID).
			Return(oldUser, nil).
			Times(1)

		// Проверка, что новый email не занят
		mockRepo.EXPECT().
			GetByEmail(ctx, newEmail).
			Return(nil, errors.New("not found")).
			Times(1)

		// Обновление пользователя
		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *models.User) error {
				if user.Email != newEmail {
					t.Errorf("Ожидался email '%s', получен '%s'", newEmail, user.Email)
				}
				return nil
			}).
			Times(1)

		// Получение обновленного пользователя
		updatedUser := &models.User{
			ID:        userID,
			Email:     newEmail,
			Nickname:  "testuser",
			CreatedAt: oldUser.CreatedAt,
			UpdatedAt: time.Now(),
		}

		mockRepo.EXPECT().
			GetByID(ctx, userID).
			Return(updatedUser, nil).
			Times(1)

		mockRepo.EXPECT().
			GetRoles(ctx, userID).
			Return(roles, nil).
			Times(1)

		result, err := userService.UpdateUser(ctx, userID, updateData)

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if result == nil {
			t.Fatal("Ожидались данные пользователя, получен nil")
		}

		if result.Email != newEmail {
			t.Errorf("Ожидался email '%s', получен '%s'", newEmail, result.Email)
		}
	})

	t.Run("email уже используется", func(t *testing.T) {
		oldUser := &models.User{
			ID:       userID,
			Email:    "old@example.com",
			Nickname: "testuser",
		}

		newEmail := "existing@example.com"
		updateData := service.UserUpdateData{
			Email: &newEmail,
		}

		existingUser := &models.User{
			ID:    2, // Другой пользователь
			Email: newEmail,
		}

		mockRepo.EXPECT().
			GetByID(ctx, userID).
			Return(oldUser, nil).
			Times(1)

		mockRepo.EXPECT().
			GetByEmail(ctx, newEmail).
			Return(existingUser, nil).
			Times(1)

		result, err := userService.UpdateUser(ctx, userID, updateData)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if result != nil {
			t.Fatal("Ожидался nil результат при ошибке")
		}

		expectedErrMsg := "указанный email уже используется"
		if err.Error() != expectedErrMsg {
			t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedErrMsg, err.Error())
		}
	})

	t.Run("успешное обновление пароля", func(t *testing.T) {
		oldUser := &models.User{
			ID:        userID,
			Email:     "test@example.com",
			Nickname:  "testuser",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}

		newPassword := "newpassword123"
		updateData := service.UserUpdateData{
			Password: &newPassword,
		}

		roles := []models.RoleEntity{
			{ID: 1, Name: "user"},
		}

		mockRepo.EXPECT().
			GetByID(ctx, userID).
			Return(oldUser, nil).
			Times(1)

		mockRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *models.User) error {
				if user.PasswordHash == "" {
					t.Error("Ожидался хешированный пароль")
				}
				return nil
			}).
			Times(1)

		updatedUser := &models.User{
			ID:        userID,
			Email:     "test@example.com",
			Nickname:  "testuser",
			CreatedAt: oldUser.CreatedAt,
			UpdatedAt: time.Now(),
		}

		mockRepo.EXPECT().
			GetByID(ctx, userID).
			Return(updatedUser, nil).
			Times(1)

		mockRepo.EXPECT().
			GetRoles(ctx, userID).
			Return(roles, nil).
			Times(1)

		result, err := userService.UpdateUser(ctx, userID, updateData)

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if result == nil {
			t.Fatal("Ожидались данные пользователя, получен nil")
		}
	})

	t.Run("ошибка получения пользователя", func(t *testing.T) {
		expectedErr := errors.New("user not found")
		mockRepo.EXPECT().
			GetByID(ctx, userID).
			Return(nil, expectedErr).
			Times(1)

		updateData := service.UserUpdateData{}

		result, err := userService.UpdateUser(ctx, userID, updateData)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if result != nil {
			t.Fatal("Ожидался nil результат при ошибке")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockUser(ctrl)
	userService := NewUserService(mockRepo)

	ctx := context.Background()
	userID := int64(1)

	t.Run("успешное удаление пользователя", func(t *testing.T) {
		mockRepo.EXPECT().
			Delete(ctx, userID).
			Return(nil).
			Times(1)

		err := userService.DeleteUser(ctx, userID)

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}
	})

	t.Run("ошибка удаления пользователя", func(t *testing.T) {
		expectedErr := errors.New("delete error")
		mockRepo.EXPECT().
			Delete(ctx, userID).
			Return(expectedErr).
			Times(1)

		err := userService.DeleteUser(ctx, userID)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
	})
}

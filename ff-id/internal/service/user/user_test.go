package user

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"github.com/ivasnev/FinFlow/ff-id/internal/repository/mock"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
)

func TestUserService_GetUserByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockAvatarRepo := mock.NewMockAvatar(ctrl)
	userService := NewUserService(mockUserRepo, mockAvatarRepo)

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

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(user, nil).
			Times(1)

		result, err := userService.GetUserByID(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.ID)
		assert.Equal(t, "test@example.com", result.Email)
		assert.Equal(t, "testuser", result.Nickname)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		expectedErr := errors.New("user not found")
		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := userService.GetUserByID(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ошибка получения пользователя")
	})

	t.Run("пользователь с опциональными полями", func(t *testing.T) {
		phone := "+1234567890"
		name := "Test User"
		birthdate := time.Now().Add(-30 * 365 * 24 * time.Hour)
		avatarID := uuid.New()

		user := &models.User{
			ID:        userID,
			Email:     "test@example.com",
			Nickname:  "testuser",
			Phone:     sql.NullString{String: phone, Valid: true},
			Name:      sql.NullString{String: name, Valid: true},
			Birthdate: sql.NullTime{Time: birthdate, Valid: true},
			AvatarID:  uuid.NullUUID{UUID: avatarID, Valid: true},
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(user, nil).
			Times(1)

		result, err := userService.GetUserByID(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Phone)
		assert.Equal(t, phone, *result.Phone)
		assert.NotNil(t, result.Name)
		assert.Equal(t, name, *result.Name)
		assert.NotNil(t, result.Birthdate)
		assert.Equal(t, birthdate.Unix(), *result.Birthdate)
		assert.NotNil(t, result.AvatarID)
		assert.Equal(t, avatarID, *result.AvatarID)
	})
}

func TestUserService_GetUsersByIds(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockAvatarRepo := mock.NewMockAvatar(ctrl)
	userService := NewUserService(mockUserRepo, mockAvatarRepo)

	ctx := context.Background()
	ids := []int64{1, 2, 3}

	t.Run("успешное получение пользователей", func(t *testing.T) {
		users := []*models.User{
			{ID: 1, Email: "user1@example.com", Nickname: "user1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: 2, Email: "user2@example.com", Nickname: "user2", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			{ID: 3, Email: "user3@example.com", Nickname: "user3", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		}

		mockUserRepo.EXPECT().
			GetByIDs(ctx, ids).
			Return(users, nil).
			Times(1)

		result, err := userService.GetUsersByIds(ctx, ids)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 3, len(result))
		assert.Equal(t, int64(1), result[0].ID)
		assert.Equal(t, "user1@example.com", result[0].Email)
	})

	t.Run("ошибка получения пользователей", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockUserRepo.EXPECT().
			GetByIDs(ctx, ids).
			Return(nil, expectedErr).
			Times(1)

		result, err := userService.GetUsersByIds(ctx, ids)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ошибка получения пользователей")
	})
}

func TestUserService_GetUserByNickname(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockAvatarRepo := mock.NewMockAvatar(ctrl)
	userService := NewUserService(mockUserRepo, mockAvatarRepo)

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

		mockUserRepo.EXPECT().
			GetByNickname(ctx, nickname).
			Return(user, nil).
			Times(1)

		result, err := userService.GetUserByNickname(ctx, nickname)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, nickname, result.Nickname)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		expectedErr := errors.New("user not found")
		mockUserRepo.EXPECT().
			GetByNickname(ctx, nickname).
			Return(nil, expectedErr).
			Times(1)

		result, err := userService.GetUserByNickname(ctx, nickname)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ошибка получения пользователя")
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockAvatarRepo := mock.NewMockAvatar(ctrl)
	userService := NewUserService(mockUserRepo, mockAvatarRepo)

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
		updateReq := service.UpdateUserRequest{
			Email: &newEmail,
		}

		updatedUser := &models.User{
			ID:        userID,
			Email:     newEmail,
			Nickname:  "testuser",
			CreatedAt: oldUser.CreatedAt,
			UpdatedAt: time.Now(),
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(oldUser, nil).
			Times(1)

		mockUserRepo.EXPECT().
			GetByEmail(ctx, newEmail).
			Return(nil, errors.New("not found")).
			Times(1)

		mockUserRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *models.User) error {
				assert.Equal(t, newEmail, user.Email)
				return nil
			}).
			Times(1)

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(updatedUser, nil).
			Times(1)

		result, err := userService.UpdateUser(ctx, userID, updateReq)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newEmail, result.Email)
	})

	t.Run("email уже используется", func(t *testing.T) {
		oldUser := &models.User{
			ID:       userID,
			Email:    "old@example.com",
			Nickname: "testuser",
		}

		newEmail := "existing@example.com"
		updateReq := service.UpdateUserRequest{
			Email: &newEmail,
		}

		existingUser := &models.User{
			ID:    2,
			Email: newEmail,
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(oldUser, nil).
			Times(1)

		mockUserRepo.EXPECT().
			GetByEmail(ctx, newEmail).
			Return(existingUser, nil).
			Times(1)

		result, err := userService.UpdateUser(ctx, userID, updateReq)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "указанный email уже используется", err.Error())
	})

	t.Run("успешное обновление никнейма", func(t *testing.T) {
		oldUser := &models.User{
			ID:        userID,
			Email:     "test@example.com",
			Nickname:  "oldnick",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}

		newNickname := "newnick"
		updateReq := service.UpdateUserRequest{
			Nickname: &newNickname,
		}

		updatedUser := &models.User{
			ID:        userID,
			Email:     "test@example.com",
			Nickname:  newNickname,
			CreatedAt: oldUser.CreatedAt,
			UpdatedAt: time.Now(),
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(oldUser, nil).
			Times(1)

		mockUserRepo.EXPECT().
			GetByNickname(ctx, newNickname).
			Return(nil, errors.New("not found")).
			Times(1)

		mockUserRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *models.User) error {
				assert.Equal(t, newNickname, user.Nickname)
				return nil
			}).
			Times(1)

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(updatedUser, nil).
			Times(1)

		result, err := userService.UpdateUser(ctx, userID, updateReq)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newNickname, result.Nickname)
	})
}

func TestUserService_ChangeAvatar(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockAvatarRepo := mock.NewMockAvatar(ctrl)
	userService := NewUserService(mockUserRepo, mockAvatarRepo)

	ctx := context.Background()
	userID := int64(1)
	fileID := uuid.New()

	t.Run("успешное изменение аватара", func(t *testing.T) {
		user := &models.User{
			ID:        userID,
			Email:     "test@example.com",
			Nickname:  "testuser",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(user, nil).
			Times(1)

		mockAvatarRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, avatar *models.UserAvatar) error {
				assert.Equal(t, userID, avatar.UserID)
				assert.Equal(t, fileID, avatar.FileID)
				return nil
			}).
			Times(1)

		mockUserRepo.EXPECT().
			Update(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *models.User) error {
				assert.True(t, user.AvatarID.Valid)
				return nil
			}).
			Times(1)

		err := userService.ChangeAvatar(ctx, userID, fileID)

		assert.NoError(t, err)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		expectedErr := errors.New("user not found")
		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(nil, expectedErr).
			Times(1)

		err := userService.ChangeAvatar(ctx, userID, fileID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка получения пользователя")
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockAvatarRepo := mock.NewMockAvatar(ctrl)
	userService := NewUserService(mockUserRepo, mockAvatarRepo)

	ctx := context.Background()
	userID := int64(1)

	t.Run("успешное удаление пользователя", func(t *testing.T) {
		mockUserRepo.EXPECT().
			Delete(ctx, userID).
			Return(nil).
			Times(1)

		err := userService.DeleteUser(ctx, userID)

		assert.NoError(t, err)
	})

	t.Run("ошибка удаления пользователя", func(t *testing.T) {
		expectedErr := errors.New("delete error")
		mockUserRepo.EXPECT().
			Delete(ctx, userID).
			Return(expectedErr).
			Times(1)

		err := userService.DeleteUser(ctx, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestUserService_RegisterUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockAvatarRepo := mock.NewMockAvatar(ctrl)
	userService := NewUserService(mockUserRepo, mockAvatarRepo)

	ctx := context.Background()
	userID := int64(1)

	t.Run("успешная регистрация пользователя", func(t *testing.T) {
		req := &service.RegisterUserRequest{
			Email:    "test@example.com",
			Nickname: "testuser",
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(nil, errors.New("not found")).
			Times(1)

		mockUserRepo.EXPECT().
			GetByEmail(ctx, req.Email).
			Return(nil, errors.New("not found")).
			Times(1)

		mockUserRepo.EXPECT().
			GetByNickname(ctx, req.Nickname).
			Return(nil, errors.New("not found")).
			Times(1)

		mockUserRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *models.User) error {
				assert.Equal(t, userID, user.ID)
				assert.Equal(t, req.Email, user.Email)
				assert.Equal(t, req.Nickname, user.Nickname)
				return nil
			}).
			Times(1)

		result, err := userService.RegisterUser(ctx, userID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.ID)
		assert.Equal(t, req.Email, result.Email)
		assert.Equal(t, req.Nickname, result.Nickname)
	})

	t.Run("пользователь с таким ID уже существует", func(t *testing.T) {
		req := &service.RegisterUserRequest{
			Email:    "test@example.com",
			Nickname: "testuser",
		}

		existingUser := &models.User{
			ID:       userID,
			Email:    "existing@example.com",
			Nickname: "existinguser",
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(existingUser, nil).
			Times(1)

		result, err := userService.RegisterUser(ctx, userID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "уже существует")
	})

	t.Run("email уже используется", func(t *testing.T) {
		req := &service.RegisterUserRequest{
			Email:    "existing@example.com",
			Nickname: "testuser",
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(nil, errors.New("not found")).
			Times(1)

		existingUser := &models.User{
			ID:    2,
			Email: req.Email,
		}

		mockUserRepo.EXPECT().
			GetByEmail(ctx, req.Email).
			Return(existingUser, nil).
			Times(1)

		result, err := userService.RegisterUser(ctx, userID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "указанный email уже используется", err.Error())
	})

	t.Run("никнейм уже используется", func(t *testing.T) {
		req := &service.RegisterUserRequest{
			Email:    "test@example.com",
			Nickname: "existinguser",
		}

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(nil, errors.New("not found")).
			Times(1)

		mockUserRepo.EXPECT().
			GetByEmail(ctx, req.Email).
			Return(nil, errors.New("not found")).
			Times(1)

		existingUser := &models.User{
			ID:       2,
			Nickname: req.Nickname,
		}

		mockUserRepo.EXPECT().
			GetByNickname(ctx, req.Nickname).
			Return(existingUser, nil).
			Times(1)

		result, err := userService.RegisterUser(ctx, userID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "указанный никнейм уже используется", err.Error())
	})
}


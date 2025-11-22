package user

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ivasnev/FinFlow/ff-split/internal/adapters"
	adaptersMock "github.com/ivasnev/FinFlow/ff-split/internal/adapters/mock"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	repositoryMock "github.com/ivasnev/FinFlow/ff-split/internal/repository/mock"
	"github.com/stretchr/testify/assert"
)

func TestUserService_CreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repositoryMock.NewMockUser(ctrl)
	mockIDAdapter := adaptersMock.NewMockIDAdapter(ctrl)
	userService := NewUserService(mockUserRepo, mockIDAdapter)

	ctx := context.Background()
	userID := int64(1)

	t.Run("успешное создание пользователя", func(t *testing.T) {
		user := &models.User{
			UserID:         &userID,
			NicknameCashed: "testuser",
			NameCashed:     "Test User",
			IsDummy:        false,
		}

		createdUser := &models.User{
			ID:             1,
			UserID:         &userID,
			NicknameCashed: "testuser",
			NameCashed:     "Test User",
			IsDummy:        false,
		}

		mockUserRepo.EXPECT().
			GetByExternalUserID(ctx, userID).
			Return(nil, repository.ErrUserNotFound).
			Times(1)

		mockUserRepo.EXPECT().
			Create(ctx, user).
			Return(createdUser, nil).
			Times(1)

		result, err := userService.CreateUser(ctx, user)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.ID)
		assert.Equal(t, userID, *result.UserID)
	})

	t.Run("пользователь уже существует", func(t *testing.T) {
		user := &models.User{
			UserID:         &userID,
			NicknameCashed: "testuser",
		}

		existingUser := &models.User{
			ID:     1,
			UserID: &userID,
		}

		mockUserRepo.EXPECT().
			GetByExternalUserID(ctx, userID).
			Return(existingUser, nil).
			Times(1)

		result, err := userService.CreateUser(ctx, user)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, existingUser.ID, result.ID)
	})
}

func TestUserService_CreateDummyUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repositoryMock.NewMockUser(ctrl)
	mockIDAdapter := adaptersMock.NewMockIDAdapter(ctrl)
	userService := NewUserService(mockUserRepo, mockIDAdapter)

	ctx := context.Background()
	eventID := int64(1)
	name := "Dummy User"

	t.Run("успешное создание dummy пользователя", func(t *testing.T) {
		createdUser := &models.User{
			ID:         1,
			NameCashed: name,
			IsDummy:    true,
		}

		mockUserRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *models.User) (*models.User, error) {
				assert.Equal(t, name, user.NameCashed)
				assert.True(t, user.IsDummy)
				return createdUser, nil
			}).
			Times(1)

		mockUserRepo.EXPECT().
			AddUserToEvent(ctx, createdUser.ID, eventID).
			Return(nil).
			Times(1)

		result, err := userService.CreateDummyUser(ctx, name, eventID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, name, result.NameCashed)
		assert.True(t, result.IsDummy)
	})

	t.Run("ошибка создания dummy пользователя", func(t *testing.T) {
		expectedErr := errors.New("creation error")
		mockUserRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil, expectedErr).
			Times(1)

		result, err := userService.CreateDummyUser(ctx, name, eventID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ошибка при создании dummy пользователя")
	})
}

func TestUserService_GetUserByInternalUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repositoryMock.NewMockUser(ctrl)
	mockIDAdapter := adaptersMock.NewMockIDAdapter(ctrl)
	userService := NewUserService(mockUserRepo, mockIDAdapter)

	ctx := context.Background()
	internalID := int64(1)

	t.Run("успешное получение пользователя", func(t *testing.T) {
		user := &models.User{
			ID:     internalID,
			UserID: func() *int64 { id := int64(100); return &id }(),
		}

		mockUserRepo.EXPECT().
			GetByInternalUserID(ctx, internalID).
			Return(user, nil).
			Times(1)

		result, err := userService.GetUserByInternalUserID(ctx, internalID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, internalID, result.ID)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		expectedErr := errors.New("user not found")
		mockUserRepo.EXPECT().
			GetByInternalUserID(ctx, internalID).
			Return(nil, expectedErr).
			Times(1)

		result, err := userService.GetUserByInternalUserID(ctx, internalID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ошибка при получении пользователя")
	})
}

func TestUserService_GetUserByExternalUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repositoryMock.NewMockUser(ctrl)
	mockIDAdapter := adaptersMock.NewMockIDAdapter(ctrl)
	userService := NewUserService(mockUserRepo, mockIDAdapter)

	ctx := context.Background()
	externalID := int64(100)

	t.Run("успешное получение пользователя из БД", func(t *testing.T) {
		user := &models.User{
			ID:     1,
			UserID: &externalID,
		}

		mockUserRepo.EXPECT().
			GetByExternalUserID(ctx, externalID).
			Return(user, nil).
			Times(1)

		result, err := userService.GetUserByExternalUserID(ctx, externalID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, externalID, *result.UserID)
	})

	t.Run("пользователь не найден, синхронизация с ID-сервисом", func(t *testing.T) {
		userID := int64(100)
		userInfo := &adapters.UserDTO{
			ID:       userID,
			Nickname: "testuser",
			Name:     func() *string { s := "Test User"; return &s }(),
		}

		syncedUser := &models.User{
			ID:             1,
			UserID:         &userID,
			NicknameCashed: "testuser",
			NameCashed:     "Test User",
		}

		mockUserRepo.EXPECT().
			GetByExternalUserID(ctx, externalID).
			Return(nil, repository.ErrUserNotFound).
			Times(1)

		mockIDAdapter.EXPECT().
			GetUserByID(ctx, externalID).
			Return(userInfo, nil).
			Times(1)

		mockUserRepo.EXPECT().
			CreateOrUpdate(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, user *models.User) error {
				assert.Equal(t, userID, *user.UserID)
				assert.Equal(t, "testuser", user.NicknameCashed)
				return nil
			}).
			Times(1)

		// Мок возвращает пользователя после CreateOrUpdate
		mockUserRepo.EXPECT().
			GetByExternalUserID(ctx, externalID).
			Return(syncedUser, nil).
			AnyTimes()

		result, err := userService.GetUserByExternalUserID(ctx, externalID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestUserService_GetUsersByExternalUserIDs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repositoryMock.NewMockUser(ctrl)
	mockIDAdapter := adaptersMock.NewMockIDAdapter(ctrl)
	userService := NewUserService(mockUserRepo, mockIDAdapter)

	ctx := context.Background()
	userIDs := []int64{100, 200}

	t.Run("успешное получение пользователей", func(t *testing.T) {
		users := []models.User{
			{ID: 1, UserID: &userIDs[0]},
			{ID: 2, UserID: &userIDs[1]},
		}

		mockUserRepo.EXPECT().
			GetByExternalUserIDs(ctx, userIDs).
			Return(users, nil).
			Times(1)

		result, err := userService.GetUsersByExternalUserIDs(ctx, userIDs)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
	})

	t.Run("ошибка получения пользователей", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockUserRepo.EXPECT().
			GetByExternalUserIDs(ctx, userIDs).
			Return(nil, expectedErr).
			Times(1)

		result, err := userService.GetUsersByExternalUserIDs(ctx, userIDs)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ошибка при получении пользователей")
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repositoryMock.NewMockUser(ctrl)
	mockIDAdapter := adaptersMock.NewMockIDAdapter(ctrl)
	userService := NewUserService(mockUserRepo, mockIDAdapter)

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
		assert.Contains(t, err.Error(), "ошибка при удалении пользователя")
	})
}

func TestUserService_AddUserToEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repositoryMock.NewMockUser(ctrl)
	mockIDAdapter := adaptersMock.NewMockIDAdapter(ctrl)
	userService := NewUserService(mockUserRepo, mockIDAdapter)

	ctx := context.Background()
	userID := int64(1)
	eventID := int64(100)

	t.Run("успешное добавление пользователя в мероприятие", func(t *testing.T) {
		mockUserRepo.EXPECT().
			AddUserToEvent(ctx, userID, eventID).
			Return(nil).
			Times(1)

		err := userService.AddUserToEvent(ctx, userID, eventID)

		assert.NoError(t, err)
	})

	t.Run("ошибка добавления пользователя в мероприятие", func(t *testing.T) {
		expectedErr := errors.New("add error")
		mockUserRepo.EXPECT().
			AddUserToEvent(ctx, userID, eventID).
			Return(expectedErr).
			Times(1)

		err := userService.AddUserToEvent(ctx, userID, eventID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при добавлении пользователя в мероприятие")
	})
}

func TestUserService_RemoveUserFromEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := repositoryMock.NewMockUser(ctrl)
	mockIDAdapter := adaptersMock.NewMockIDAdapter(ctrl)
	userService := NewUserService(mockUserRepo, mockIDAdapter)

	ctx := context.Background()
	userID := int64(1)
	eventID := int64(100)

	t.Run("успешное удаление пользователя из мероприятия", func(t *testing.T) {
		mockUserRepo.EXPECT().
			RemoveUserFromEvent(ctx, userID, eventID).
			Return(nil).
			Times(1)

		err := userService.RemoveUserFromEvent(ctx, userID, eventID)

		assert.NoError(t, err)
	})

	t.Run("ошибка удаления пользователя из мероприятия", func(t *testing.T) {
		expectedErr := errors.New("remove error")
		mockUserRepo.EXPECT().
			RemoveUserFromEvent(ctx, userID, eventID).
			Return(expectedErr).
			Times(1)

		err := userService.RemoveUserFromEvent(ctx, userID, eventID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при удалении пользователя из мероприятия")
	})
}

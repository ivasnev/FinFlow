package icon

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/mock"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

func TestIconService_GetIcons(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIcon(ctrl)
	iconService := NewIconService(mockRepo)

	ctx := context.Background()

	t.Run("успешное получение иконок", func(t *testing.T) {
		icons := []models.Icon{
			{ID: 1, Name: "Icon 1", FileUUID: "uuid-1"},
			{ID: 2, Name: "Icon 2", FileUUID: "uuid-2"},
		}

		mockRepo.EXPECT().
			GetIcons().
			Return(icons, nil).
			Times(1)

		result, err := iconService.GetIcons(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "Icon 1", result[0].Name)
		assert.Equal(t, "uuid-1", result[0].FileUUID)
	})

	t.Run("ошибка получения иконок", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetIcons().
			Return(nil, expectedErr).
			Times(1)

		result, err := iconService.GetIcons(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestIconService_GetIconByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIcon(ctrl)
	iconService := NewIconService(mockRepo)

	ctx := context.Background()
	iconID := uint(1)

	t.Run("успешное получение иконки", func(t *testing.T) {
		icon := &models.Icon{
			ID:       1,
			Name:     "Test Icon",
			FileUUID: "test-uuid",
		}

		mockRepo.EXPECT().
			GetIconByID(iconID).
			Return(icon, nil).
			Times(1)

		result, err := iconService.GetIconByID(ctx, iconID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, "Test Icon", result.Name)
		assert.Equal(t, "test-uuid", result.FileUUID)
	})

	t.Run("иконка не найдена", func(t *testing.T) {
		expectedErr := errors.New("icon not found")
		mockRepo.EXPECT().
			GetIconByID(iconID).
			Return(nil, expectedErr).
			Times(1)

		result, err := iconService.GetIconByID(ctx, iconID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestIconService_CreateIcon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIcon(ctrl)
	iconService := NewIconService(mockRepo)

	ctx := context.Background()
	iconDTO := &service.IconFullDTO{
		Name:     "New Icon",
		FileUUID: "new-uuid",
	}

	t.Run("успешное создание иконки", func(t *testing.T) {
		mockRepo.EXPECT().
			CreateIcon(gomock.Any()).
			DoAndReturn(func(icon *models.Icon) error {
				icon.ID = 1
				return nil
			}).
			Times(1)

		result, err := iconService.CreateIcon(ctx, iconDTO)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, "New Icon", result.Name)
	})

	t.Run("ошибка создания иконки", func(t *testing.T) {
		expectedErr := errors.New("creation error")
		mockRepo.EXPECT().
			CreateIcon(gomock.Any()).
			Return(expectedErr).
			Times(1)

		result, err := iconService.CreateIcon(ctx, iconDTO)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestIconService_UpdateIcon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIcon(ctrl)
	iconService := NewIconService(mockRepo)

	ctx := context.Background()
	iconID := uint(1)
	iconDTO := &service.IconFullDTO{
		ID:       iconID,
		Name:     "Updated Icon",
		FileUUID: "updated-uuid",
	}

	t.Run("успешное обновление иконки", func(t *testing.T) {
		mockRepo.EXPECT().
			UpdateIcon(gomock.Any()).
			DoAndReturn(func(icon *models.Icon) error {
				assert.Equal(t, 1, icon.ID)
				assert.Equal(t, "Updated Icon", icon.Name)
				return nil
			}).
			Times(1)

		result, err := iconService.UpdateIcon(ctx, iconID, iconDTO)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, iconID, result.ID)
		assert.Equal(t, "Updated Icon", result.Name)
	})

	t.Run("ошибка обновления иконки", func(t *testing.T) {
		expectedErr := errors.New("update error")
		mockRepo.EXPECT().
			UpdateIcon(gomock.Any()).
			Return(expectedErr).
			Times(1)

		result, err := iconService.UpdateIcon(ctx, iconID, iconDTO)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestIconService_DeleteIcon(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIcon(ctrl)
	iconService := NewIconService(mockRepo)

	ctx := context.Background()
	iconID := uint(1)

	t.Run("успешное удаление иконки", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteIcon(iconID).
			Return(nil).
			Times(1)

		err := iconService.DeleteIcon(ctx, iconID)

		assert.NoError(t, err)
	})

	t.Run("ошибка удаления иконки", func(t *testing.T) {
		expectedErr := errors.New("delete error")
		mockRepo.EXPECT().
			DeleteIcon(iconID).
			Return(expectedErr).
			Times(1)

		err := iconService.DeleteIcon(ctx, iconID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}


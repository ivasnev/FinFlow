package activity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/mock"
)

func TestActivityService_GetActivitiesByEventID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockActivity(ctrl)
	activityService := NewActivityService(mockRepo)

	ctx := context.Background()
	eventID := int64(1)

	t.Run("успешное получение активностей", func(t *testing.T) {
		activities := []models.Activity{
			{
				ID:          1,
				EventID:     &eventID,
				Description: "Activity 1",
				IconID:      1,
				CreatedAt:   time.Now(),
			},
			{
				ID:          2,
				EventID:     &eventID,
				Description: "Activity 2",
				IconID:      2,
				CreatedAt:   time.Now(),
			},
		}

		mockRepo.EXPECT().
			GetByEventID(ctx, eventID).
			Return(activities, nil).
			Times(1)

		result, err := activityService.GetActivitiesByEventID(ctx, eventID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "Activity 1", result[0].Description)
	})

	t.Run("ошибка получения активностей", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetByEventID(ctx, eventID).
			Return(nil, expectedErr).
			Times(1)

		result, err := activityService.GetActivitiesByEventID(ctx, eventID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestActivityService_GetActivityByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockActivity(ctrl)
	activityService := NewActivityService(mockRepo)

	ctx := context.Background()
	activityID := 1

	t.Run("успешное получение активности", func(t *testing.T) {
		activity := &models.Activity{
			ID:          activityID,
			Description: "Test Activity",
			IconID:      1,
			CreatedAt:   time.Now(),
		}

		mockRepo.EXPECT().
			GetByID(ctx, activityID).
			Return(activity, nil).
			Times(1)

		result, err := activityService.GetActivityByID(ctx, activityID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, activityID, result.ID)
		assert.Equal(t, "Test Activity", result.Description)
	})

	t.Run("активность не найдена", func(t *testing.T) {
		expectedErr := errors.New("activity not found")
		mockRepo.EXPECT().
			GetByID(ctx, activityID).
			Return(nil, expectedErr).
			Times(1)

		result, err := activityService.GetActivityByID(ctx, activityID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestActivityService_CreateActivity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockActivity(ctrl)
	activityService := NewActivityService(mockRepo)

	ctx := context.Background()
	eventID := int64(1)
	userID := int64(100)

	t.Run("успешное создание активности", func(t *testing.T) {
		activity := &models.Activity{
			EventID:     &eventID,
			UserID:      &userID,
			Description: "New Activity",
			IconID:      1,
		}

		createdActivity := &models.Activity{
			ID:          1,
			EventID:     &eventID,
			UserID:      &userID,
			Description: "New Activity",
			IconID:      1,
			CreatedAt:   time.Now(),
		}

		mockRepo.EXPECT().
			Create(ctx, activity).
			Return(createdActivity, nil).
			Times(1)

		result, err := activityService.CreateActivity(ctx, activity)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "New Activity", result.Description)
	})

	t.Run("ошибка создания активности", func(t *testing.T) {
		activity := &models.Activity{
			Description: "New Activity",
			IconID:      1,
		}

		expectedErr := errors.New("creation error")
		mockRepo.EXPECT().
			Create(ctx, activity).
			Return(nil, expectedErr).
			Times(1)

		result, err := activityService.CreateActivity(ctx, activity)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestActivityService_UpdateActivity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockActivity(ctrl)
	activityService := NewActivityService(mockRepo)

	ctx := context.Background()
	activityID := 1

	t.Run("успешное обновление активности", func(t *testing.T) {
		activity := &models.Activity{
			ID:          activityID,
			Description: "Updated Activity",
			IconID:      2,
		}

		updatedActivity := &models.Activity{
			ID:          activityID,
			Description: "Updated Activity",
			IconID:      2,
			CreatedAt:   time.Now(),
		}

		mockRepo.EXPECT().
			Update(ctx, activityID, activity).
			Return(updatedActivity, nil).
			Times(1)

		result, err := activityService.UpdateActivity(ctx, activityID, activity)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, activityID, result.ID)
		assert.Equal(t, "Updated Activity", result.Description)
	})

	t.Run("ошибка обновления активности", func(t *testing.T) {
		activity := &models.Activity{
			ID:          activityID,
			Description: "Updated Activity",
		}

		expectedErr := errors.New("update error")
		mockRepo.EXPECT().
			Update(ctx, activityID, activity).
			Return(nil, expectedErr).
			Times(1)

		result, err := activityService.UpdateActivity(ctx, activityID, activity)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestActivityService_DeleteActivity(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockActivity(ctrl)
	activityService := NewActivityService(mockRepo)

	ctx := context.Background()
	activityID := 1

	t.Run("успешное удаление активности", func(t *testing.T) {
		mockRepo.EXPECT().
			Delete(ctx, activityID).
			Return(nil).
			Times(1)

		err := activityService.DeleteActivity(ctx, activityID)

		assert.NoError(t, err)
	})

	t.Run("ошибка удаления активности", func(t *testing.T) {
		expectedErr := errors.New("delete error")
		mockRepo.EXPECT().
			Delete(ctx, activityID).
			Return(expectedErr).
			Times(1)

		err := activityService.DeleteActivity(ctx, activityID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}


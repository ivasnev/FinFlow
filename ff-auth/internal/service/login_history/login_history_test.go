package login_history

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/mock"
	"github.com/stretchr/testify/assert"
)

func TestLoginHistoryService_GetUserLoginHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockLoginHistory(ctrl)
	loginHistoryService := NewLoginHistoryService(mockRepo)

	ctx := context.Background()
	userID := int64(1)
	limit := 10
	offset := 0

	t.Run("успешное получение истории входов", func(t *testing.T) {
		history := []models.LoginHistory{
			{
				ID:        1,
				UserID:    userID,
				IPAddress: "192.168.1.1",
				UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
				CreatedAt: time.Now().Add(-2 * time.Hour),
			},
			{
				ID:        2,
				UserID:    userID,
				IPAddress: "192.168.1.2",
				UserAgent: "Chrome/91.0.4472.124 Safari/537.36",
				CreatedAt: time.Now().Add(-1 * time.Hour),
			},
			{
				ID:        3,
				UserID:    userID,
				IPAddress: "10.0.0.1",
				UserAgent: "", // Пустой UserAgent
				CreatedAt: time.Now().Add(-30 * time.Minute),
			},
		}

		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID, limit, offset).
			Return(history, nil).
			Times(1)

		result, err := loginHistoryService.GetUserLoginHistory(ctx, userID, limit, offset)

		assert.NoError(t, err)
		assert.Equal(t, 3, len(result))
		assert.Equal(t, 1, result[0].Id)
		assert.Equal(t, "192.168.1.1", result[0].IpAddress)
		assert.NotNil(t, result[0].UserAgent)
		assert.Equal(t, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36", *result[0].UserAgent)
		assert.Equal(t, 2, result[1].Id)
		assert.Equal(t, "192.168.1.2", result[1].IpAddress)
		assert.NotNil(t, result[1].UserAgent)
		assert.Equal(t, "Chrome/91.0.4472.124 Safari/537.36", *result[1].UserAgent)
		assert.Equal(t, 3, result[2].Id)
		assert.Equal(t, "10.0.0.1", result[2].IpAddress)
		assert.Nil(t, result[2].UserAgent)
	})

	t.Run("пустая история входов", func(t *testing.T) {
		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID, limit, offset).
			Return([]models.LoginHistory{}, nil).
			Times(1)

		result, err := loginHistoryService.GetUserLoginHistory(ctx, userID, limit, offset)

		assert.NoError(t, err)
		assert.Equal(t, 0, len(result))
	})

	t.Run("ошибка репозитория", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID, limit, offset).
			Return(nil, expectedErr).
			Times(1)

		result, err := loginHistoryService.GetUserLoginHistory(ctx, userID, limit, offset)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("разные лимиты и оффсеты", func(t *testing.T) {
		history := []models.LoginHistory{
			{
				ID:        1,
				UserID:    userID,
				IPAddress: "192.168.1.1",
				UserAgent: "Mozilla/5.0",
				CreatedAt: time.Now(),
			},
		}

		testCases := []struct {
			limit  int
			offset int
		}{
			{5, 0},
			{10, 5},
			{20, 10},
			{1, 0},
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("limit=%d,offset=%d", tc.limit, tc.offset), func(t *testing.T) {
				mockRepo.EXPECT().
					GetAllByUserID(ctx, userID, tc.limit, tc.offset).
					Return(history, nil).
					Times(1)

				result, err := loginHistoryService.GetUserLoginHistory(ctx, userID, tc.limit, tc.offset)

				assert.NoError(t, err)
				assert.Equal(t, 1, len(result))
			})
		}
	})
}

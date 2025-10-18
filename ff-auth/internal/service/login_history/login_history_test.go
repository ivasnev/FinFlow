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

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if len(result) != 3 {
			t.Fatalf("Ожидалось 3 записи, получено %d", len(result))
		}

		// Проверяем первую запись
		if result[0].Id != 1 {
			t.Errorf("Ожидался ID 1, получен %d", result[0].Id)
		}
		if result[0].IpAddress != "192.168.1.1" {
			t.Errorf("Ожидался IP '192.168.1.1', получен '%s'", result[0].IpAddress)
		}
		if result[0].UserAgent == nil {
			t.Error("Ожидался UserAgent, получен nil")
		} else if *result[0].UserAgent != "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" {
			t.Errorf("Ожидался UserAgent 'Mozilla/5.0...', получен '%s'", *result[0].UserAgent)
		}

		// Проверяем вторую запись
		if result[1].Id != 2 {
			t.Errorf("Ожидался ID 2, получен %d", result[1].Id)
		}
		if result[1].IpAddress != "192.168.1.2" {
			t.Errorf("Ожидался IP '192.168.1.2', получен '%s'", result[1].IpAddress)
		}
		if result[1].UserAgent == nil {
			t.Error("Ожидался UserAgent, получен nil")
		} else if *result[1].UserAgent != "Chrome/91.0.4472.124 Safari/537.36" {
			t.Errorf("Ожидался UserAgent 'Chrome/91.0...', получен '%s'", *result[1].UserAgent)
		}

		// Проверяем третью запись (с пустым UserAgent)
		if result[2].Id != 3 {
			t.Errorf("Ожидался ID 3, получен %d", result[2].Id)
		}
		if result[2].IpAddress != "10.0.0.1" {
			t.Errorf("Ожидался IP '10.0.0.1', получен '%s'", result[2].IpAddress)
		}
		if result[2].UserAgent != nil {
			t.Errorf("Ожидался nil UserAgent, получен '%s'", *result[2].UserAgent)
		}
	})

	t.Run("пустая история входов", func(t *testing.T) {
		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID, limit, offset).
			Return([]models.LoginHistory{}, nil).
			Times(1)

		result, err := loginHistoryService.GetUserLoginHistory(ctx, userID, limit, offset)

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if len(result) != 0 {
			t.Fatalf("Ожидалось 0 записей, получено %d", len(result))
		}
	})

	t.Run("ошибка репозитория", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID, limit, offset).
			Return(nil, expectedErr).
			Times(1)

		result, err := loginHistoryService.GetUserLoginHistory(ctx, userID, limit, offset)

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

				if err != nil {
					t.Fatalf("Ожидался успех, получена ошибка: %v", err)
				}

				if len(result) != 1 {
					t.Fatalf("Ожидалась 1 запись, получено %d", len(result))
				}
			})
		}
	})
}

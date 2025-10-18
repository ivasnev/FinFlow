package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/mock"
)

func TestSessionService_GetUserSessions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSession(ctrl)
	sessionService := NewSessionService(mockRepo)

	ctx := context.Background()
	userID := int64(1)

	t.Run("успешное получение сессий", func(t *testing.T) {
		sessionID1 := uuid.New()
		sessionID2 := uuid.New()

		sessions := []models.Session{
			{
				ID:        sessionID1,
				UserID:    userID,
				IPAddress: []string{"192.168.1.1"},
				CreatedAt: time.Now().Add(-2 * time.Hour),
				ExpiresAt: time.Now().Add(2 * time.Hour),
			},
			{
				ID:        sessionID2,
				UserID:    userID,
				IPAddress: []string{"192.168.1.2", "10.0.0.1"},
				CreatedAt: time.Now().Add(-1 * time.Hour),
				ExpiresAt: time.Now().Add(3 * time.Hour),
			},
		}

		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID).
			Return(sessions, nil).
			Times(1)

		result, err := sessionService.GetUserSessions(ctx, userID)

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("Ожидалось 2 сессии, получено %d", len(result))
		}

		// Проверяем первую сессию
		if result[0].Id != sessionID1 {
			t.Errorf("Ожидался ID %s, получен %s", sessionID1, result[0].Id)
		}
		if result[0].IpAddress != "192.168.1.1" {
			t.Errorf("Ожидался IP '192.168.1.1', получен '%s'", result[0].IpAddress)
		}

		// Проверяем вторую сессию (должен взять первый IP)
		if result[1].Id != sessionID2 {
			t.Errorf("Ожидался ID %s, получен %s", sessionID2, result[1].Id)
		}
		if result[1].IpAddress != "192.168.1.2" {
			t.Errorf("Ожидался IP '192.168.1.2', получен '%s'", result[1].IpAddress)
		}
	})

	t.Run("сессии с пустым IP", func(t *testing.T) {
		sessionID := uuid.New()
		sessions := []models.Session{
			{
				ID:        sessionID,
				UserID:    userID,
				IPAddress: []string{},
				CreatedAt: time.Now().Add(-1 * time.Hour),
				ExpiresAt: time.Now().Add(2 * time.Hour),
			},
		}

		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID).
			Return(sessions, nil).
			Times(1)

		result, err := sessionService.GetUserSessions(ctx, userID)

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if len(result) != 1 {
			t.Fatalf("Ожидалась 1 сессия, получено %d", len(result))
		}

		if result[0].IpAddress != "" {
			t.Errorf("Ожидался пустой IP, получен '%s'", result[0].IpAddress)
		}
	})

	t.Run("ошибка репозитория", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := sessionService.GetUserSessions(ctx, userID)

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

func TestSessionService_TerminateSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSession(ctrl)
	sessionService := NewSessionService(mockRepo)

	ctx := context.Background()
	userID := int64(1)
	sessionID := uuid.New()

	t.Run("успешное завершение сессии", func(t *testing.T) {
		session := &models.Session{
			ID:        sessionID,
			UserID:    userID,
			IPAddress: []string{"192.168.1.1"},
			CreatedAt: time.Now().Add(-1 * time.Hour),
			ExpiresAt: time.Now().Add(2 * time.Hour),
		}

		mockRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(session, nil).
			Times(1)

		mockRepo.EXPECT().
			Delete(ctx, sessionID).
			Return(nil).
			Times(1)

		err := sessionService.TerminateSession(ctx, sessionID, userID)

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}
	})

	t.Run("сессия не найдена", func(t *testing.T) {
		expectedErr := errors.New("session not found")
		mockRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(nil, expectedErr).
			Times(1)

		err := sessionService.TerminateSession(ctx, sessionID, userID)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
	})

	t.Run("сессия принадлежит другому пользователю", func(t *testing.T) {
		otherUserID := int64(2)
		session := &models.Session{
			ID:        sessionID,
			UserID:    otherUserID, // Другой пользователь
			IPAddress: []string{"192.168.1.1"},
			CreatedAt: time.Now().Add(-1 * time.Hour),
			ExpiresAt: time.Now().Add(2 * time.Hour),
		}

		mockRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(session, nil).
			Times(1)

		err := sessionService.TerminateSession(ctx, sessionID, userID)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		expectedErrMsg := "у вас нет прав на удаление этой сессии"
		if err.Error() != expectedErrMsg {
			t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedErrMsg, err.Error())
		}
	})

	t.Run("ошибка удаления сессии", func(t *testing.T) {
		session := &models.Session{
			ID:        sessionID,
			UserID:    userID,
			IPAddress: []string{"192.168.1.1"},
			CreatedAt: time.Now().Add(-1 * time.Hour),
			ExpiresAt: time.Now().Add(2 * time.Hour),
		}

		mockRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(session, nil).
			Times(1)

		expectedErr := errors.New("delete error")
		mockRepo.EXPECT().
			Delete(ctx, sessionID).
			Return(expectedErr).
			Times(1)

		err := sessionService.TerminateSession(ctx, sessionID, userID)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
	})
}

func TestSessionService_TerminateAllSessions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockSession(ctrl)
	sessionService := NewSessionService(mockRepo)

	ctx := context.Background()
	userID := int64(1)

	t.Run("успешное завершение всех сессий", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteAllByUserID(ctx, userID).
			Return(nil).
			Times(1)

		err := sessionService.TerminateAllSessions(ctx, userID)

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}
	})

	t.Run("ошибка завершения всех сессий", func(t *testing.T) {
		expectedErr := errors.New("delete all error")
		mockRepo.EXPECT().
			DeleteAllByUserID(ctx, userID).
			Return(expectedErr).
			Times(1)

		err := sessionService.TerminateAllSessions(ctx, userID)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
	})
}

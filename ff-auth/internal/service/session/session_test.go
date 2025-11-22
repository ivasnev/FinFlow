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
	"github.com/stretchr/testify/assert"
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

		assert.NoError(t, err)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, sessionID1, result[0].Id)
		assert.Equal(t, "192.168.1.1", result[0].IpAddress)
		assert.Equal(t, sessionID2, result[1].Id)
		assert.Equal(t, "192.168.1.2", result[1].IpAddress)
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

		assert.NoError(t, err)
		assert.Equal(t, 1, len(result))
		assert.Empty(t, result[0].IpAddress)
	})

	t.Run("ошибка репозитория", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := sessionService.GetUserSessions(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
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

		assert.NoError(t, err)
	})

	t.Run("сессия не найдена", func(t *testing.T) {
		expectedErr := errors.New("session not found")
		mockRepo.EXPECT().
			GetByID(ctx, sessionID).
			Return(nil, expectedErr).
			Times(1)

		err := sessionService.TerminateSession(ctx, sessionID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
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

		assert.Error(t, err)
		assert.Equal(t, "у вас нет прав на удаление этой сессии", err.Error())
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

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
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

		assert.NoError(t, err)
	})

	t.Run("ошибка завершения всех сессий", func(t *testing.T) {
		expectedErr := errors.New("delete all error")
		mockRepo.EXPECT().
			DeleteAllByUserID(ctx, userID).
			Return(expectedErr).
			Times(1)

		err := sessionService.TerminateAllSessions(ctx, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

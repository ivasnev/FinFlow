package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/interfaces"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// SessionService реализует интерфейс SessionService
type SessionService struct {
	sessionRepository interfaces.SessionRepository
}

// NewSessionService создает новый сервис сессий
func NewSessionService(sessionRepository interfaces.SessionRepository) *SessionService {
	return &SessionService{
		sessionRepository: sessionRepository,
	}
}

// GetAllByUserID получает все сессии пользователя
func (s *SessionService) GetAllByUserID(ctx context.Context, userID int64) ([]models.SessionInfo, error) {
	// Заглушка
	return []models.SessionInfo{}, nil
}

// TerminateSession завершает сессию
func (s *SessionService) TerminateSession(ctx context.Context, userID int64, sessionID uuid.UUID) error {
	// Заглушка
	return nil
}

// TerminateAllSessions завершает все сессии пользователя
func (s *SessionService) TerminateAllSessions(ctx context.Context, userID int64) error {
	// Заглушка
	return nil
}

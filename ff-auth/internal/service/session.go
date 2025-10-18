package service

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SessionParams представляет данные о сессии пользователя
type SessionParams struct {
	Id        uuid.UUID
	IpAddress string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Session определяет методы для работы с сессиями
type Session interface {
	// GetUserSessions получает все сессии пользователя
	GetUserSessions(ctx context.Context, userID int64) ([]SessionParams, error)

	// TerminateSession завершает сессию
	TerminateSession(ctx context.Context, sessionID uuid.UUID, userID int64) error

	// TerminateAllSessions завершает все сессии пользователя
	TerminateAllSessions(ctx context.Context, userID int64) error
}

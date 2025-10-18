package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// Session определяет методы для работы с сессиями
type Session interface {
	// Create создает новую сессию
	Create(ctx context.Context, session *models.Session) error

	// GetByID находит сессию по ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error)

	// GetByRefreshToken находит сессию по refresh-токену
	GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error)

	// GetAllByUserID получает все сессии пользователя
	GetAllByUserID(ctx context.Context, userID int64) ([]models.Session, error)

	// Delete удаляет сессию
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteAllByUserID удаляет все сессии пользователя
	DeleteAllByUserID(ctx context.Context, userID int64) error

	// DeleteExpired удаляет все истекшие сессии
	DeleteExpired(ctx context.Context) error
}

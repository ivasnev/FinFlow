package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// Avatar определяет методы для работы с аватарами
type Avatar interface {
	// Create создает новый аватар
	Create(ctx context.Context, avatar *models.UserAvatar) error

	// GetByID получает аватар по ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.UserAvatar, error)

	// GetAllByUserID получает все аватарки пользователя
	GetAllByUserID(ctx context.Context, userID int64) ([]models.UserAvatar, error)

	// Delete удаляет аватар
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteAllByUserID удаляет все аватарки пользователя
	DeleteAllByUserID(ctx context.Context, userID int64) error
}

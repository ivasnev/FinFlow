package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// UserRepositoryInterface определяет методы для работы с пользователями
type UserRepositoryInterface interface {
	// Create создает нового пользователя
	Create(ctx context.Context, user *models.User) error

	// GetByID получает пользователя по ID
	GetByID(ctx context.Context, id int64) (*models.User, error)

	// GetByEmail получает пользователя по email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// GetByNickname получает пользователя по никнейму
	GetByNickname(ctx context.Context, nickname string) (*models.User, error)

	// Update обновляет данные пользователя
	Update(ctx context.Context, user *models.User) error

	// Delete удаляет пользователя
	Delete(ctx context.Context, id int64) error
}

// AvatarRepositoryInterface определяет методы для работы с аватарами
type AvatarRepositoryInterface interface {
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

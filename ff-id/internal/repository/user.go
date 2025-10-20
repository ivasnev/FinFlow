package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// User определяет методы для работы с пользователями
type User interface {
	// Create создает нового пользователя
	Create(ctx context.Context, user *models.User) error

	// GetByID получает пользователя по ID
	GetByID(ctx context.Context, id int64) (*models.User, error)

	// GetByIDs получает пользователей по их ID
	GetByIDs(ctx context.Context, ids []int64) ([]*models.User, error)

	// GetByEmail получает пользователя по email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// GetByNickname получает пользователя по никнейму
	GetByNickname(ctx context.Context, nickname string) (*models.User, error)

	// Update обновляет данные пользователя
	Update(ctx context.Context, user *models.User) error

	// Delete удаляет пользователя
	Delete(ctx context.Context, id int64) error
}

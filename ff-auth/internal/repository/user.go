package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// User определяет методы для работы с пользователями
type User interface {
	// Create создает нового пользователя
	Create(ctx context.Context, user *models.User) error

	// GetByID находит пользователя по ID
	GetByID(ctx context.Context, id int64) (*models.User, error)

	// GetByEmail находит пользователя по email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// GetByNickname находит пользователя по никнейму
	GetByNickname(ctx context.Context, nickname string) (*models.User, error)

	// Update обновляет данные пользователя
	Update(ctx context.Context, user *models.User) error

	// Delete удаляет пользователя
	Delete(ctx context.Context, id int64) error

	// AddRole добавляет пользователю роль
	AddRole(ctx context.Context, userID int64, roleID int) error

	// RemoveRole удаляет роль у пользователя
	RemoveRole(ctx context.Context, userID int64, roleID int) error

	// GetRoles получает все роли пользователя
	GetRoles(ctx context.Context, userID int64) ([]models.RoleEntity, error)
}

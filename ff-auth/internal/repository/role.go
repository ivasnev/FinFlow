package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// Role определяет методы для работы с ролями
type Role interface {
	// GetByName находит роль по имени
	GetByName(ctx context.Context, name string) (*models.RoleEntity, error)

	// GetAll получает все роли
	GetAll(ctx context.Context) ([]models.RoleEntity, error)
}

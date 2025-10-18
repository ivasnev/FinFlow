package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// LoginHistory определяет методы для работы с историей входов
type LoginHistory interface {
	// Create создает новую запись в истории входов
	Create(ctx context.Context, history *models.LoginHistory) error

	// GetAllByUserID получает всю историю входов пользователя
	GetAllByUserID(ctx context.Context, userID int64, limit, offset int) ([]models.LoginHistory, error)
}

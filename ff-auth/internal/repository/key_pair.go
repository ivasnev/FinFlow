package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// KeyPair определяет методы для работы с ключами
type KeyPair interface {
	// Create создает новую пару ключей
	Create(ctx context.Context, keyPair *models.KeyPair) error

	// GetActive получает активную пару ключей
	GetActive(ctx context.Context) (*models.KeyPair, error)

	// GetByID получает пару ключей по ID
	GetByID(ctx context.Context, id int) (*models.KeyPair, error)

	// Update обновляет пару ключей
	Update(ctx context.Context, keyPair *models.KeyPair) error

	// SetActive устанавливает пару ключей как активную и деактивирует остальные
	SetActive(ctx context.Context, id int) error
}

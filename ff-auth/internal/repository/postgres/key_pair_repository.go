package postgres

import (
	"context"
	"errors"

	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"gorm.io/gorm"
)

// KeyPairRepository представляет реализацию репозитория для работы с ключами
type KeyPairRepository struct {
	db *gorm.DB
}

// NewKeyPairRepository создает новый экземпляр репозитория для работы с ключами
func NewKeyPairRepository(db *gorm.DB) KeyPairRepositoryInterface {
	return &KeyPairRepository{
		db: db,
	}
}

// Create создает новую пару ключей
func (r *KeyPairRepository) Create(ctx context.Context, keyPair *models.KeyPair) error {
	return r.db.WithContext(ctx).Create(keyPair).Error
}

// GetActive получает активную пару ключей
func (r *KeyPairRepository) GetActive(ctx context.Context) (*models.KeyPair, error) {
	var keyPair models.KeyPair
	result := r.db.WithContext(ctx).Where("is_active = ?", true).First(&keyPair)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // Нет активных ключей
		}
		return nil, result.Error
	}
	return &keyPair, nil
}

// GetByID получает пару ключей по ID
func (r *KeyPairRepository) GetByID(ctx context.Context, id int) (*models.KeyPair, error) {
	var keyPair models.KeyPair
	result := r.db.WithContext(ctx).First(&keyPair, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &keyPair, nil
}

// Update обновляет пару ключей
func (r *KeyPairRepository) Update(ctx context.Context, keyPair *models.KeyPair) error {
	return r.db.WithContext(ctx).Save(keyPair).Error
}

// SetActive устанавливает пару ключей как активную и деактивирует остальные
func (r *KeyPairRepository) SetActive(ctx context.Context, id int) error {
	// Транзакция для обеспечения атомарности
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Деактивируем все ключи
		if err := tx.Model(&models.KeyPair{}).Update("is_active", false).Error; err != nil {
			return err
		}

		// Активируем указанный ключ
		return tx.Model(&models.KeyPair{}).Where("id = ?", id).Update("is_active", true).Error
	})
}

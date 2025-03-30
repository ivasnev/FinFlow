package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"gorm.io/gorm"
)

// AvatarRepository реализует интерфейс для работы с аватарками в PostgreSQL через GORM
type AvatarRepository struct {
	db *gorm.DB
}

// NewAvatarRepository создает новый репозиторий аватарок
func NewAvatarRepository(db *gorm.DB) *AvatarRepository {
	return &AvatarRepository{
		db: db,
	}
}

// Create создает новую аватарку
func (r *AvatarRepository) Create(ctx context.Context, avatar *models.UserAvatar) error {
	return r.db.WithContext(ctx).Create(avatar).Error
}

// GetByID находит аватарку по ID
func (r *AvatarRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.UserAvatar, error) {
	var avatar models.UserAvatar
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&avatar).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("аватарка не найдена")
		}
		return nil, err
	}
	return &avatar, nil
}

// GetAllByUserID получает все аватарки пользователя
func (r *AvatarRepository) GetAllByUserID(ctx context.Context, userID int64) ([]models.UserAvatar, error) {
	var avatars []models.UserAvatar
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&avatars).Error
	return avatars, err
}

// Delete удаляет аватарку
func (r *AvatarRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.UserAvatar{}).Error
}

// DeleteAllByUserID удаляет все аватарки пользователя
func (r *AvatarRepository) DeleteAllByUserID(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.UserAvatar{}).Error
}

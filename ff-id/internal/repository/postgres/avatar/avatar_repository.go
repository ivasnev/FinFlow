package avatar

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"github.com/ivasnev/FinFlow/ff-id/internal/repository"
	"gorm.io/gorm"
)

// AvatarRepository реализует интерфейс repository.Avatar для работы с аватарами в PostgreSQL через GORM
type AvatarRepository struct {
	db *gorm.DB
}

// NewAvatarRepository создает новый репозиторий аватаров
func NewAvatarRepository(db *gorm.DB) repository.Avatar {
	return &AvatarRepository{
		db: db,
	}
}

// Create создает новый аватар
func (r *AvatarRepository) Create(ctx context.Context, avatar *models.UserAvatar) error {
	dbAvatar := LoadUserAvatar(avatar)
	return r.db.WithContext(ctx).Create(dbAvatar).Error
}

// GetByID получает аватар по ID
func (r *AvatarRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.UserAvatar, error) {
	var avatar UserAvatar
	err := r.db.WithContext(ctx).First(&avatar, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("аватар не найден")
		}
		return nil, err
	}
	return ExtractUserAvatar(&avatar), nil
}

// GetAllByUserID получает все аватарки пользователя
func (r *AvatarRepository) GetAllByUserID(ctx context.Context, userID int64) ([]models.UserAvatar, error) {
	var dbAvatars []UserAvatar
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&dbAvatars).Error
	if err != nil {
		return nil, err
	}

	avatars := make([]models.UserAvatar, len(dbAvatars))
	for i, dbAvatar := range dbAvatars {
		avatars[i] = *ExtractUserAvatar(&dbAvatar)
	}
	return avatars, nil
}

// Delete удаляет аватар
func (r *AvatarRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&UserAvatar{}, "id = ?", id).Error
}

// DeleteAllByUserID удаляет все аватарки пользователя
func (r *AvatarRepository) DeleteAllByUserID(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&UserAvatar{}).Error
}

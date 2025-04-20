package postgres

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// IconRepository представляет собой репозиторий для работы с иконками в PostgreSQL
type IconRepository struct {
	db *gorm.DB
}

// NewIconRepository создает новый экземпляр IconRepository
func NewIconRepository(db *gorm.DB) *IconRepository {
	return &IconRepository{
		db: db,
	}
}

// GetIcons возвращает все иконки
func (r *IconRepository) GetIcons(ctx context.Context) ([]models.Icon, error) {
	var icons []models.Icon
	if err := r.db.WithContext(ctx).Find(&icons).Error; err != nil {
		return nil, err
	}
	return icons, nil
}

// GetIconByID возвращает иконку по ID
func (r *IconRepository) GetIconByID(ctx context.Context, id string) (models.Icon, error) {
	var icon models.Icon
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&icon).Error; err != nil {
		return models.Icon{}, err
	}
	return icon, nil
}

// CreateIcon создает новую иконку
func (r *IconRepository) CreateIcon(ctx context.Context, icon models.Icon) (models.Icon, error) {
	if err := r.db.WithContext(ctx).Create(&icon).Error; err != nil {
		return models.Icon{}, err
	}
	return icon, nil
}

// UpdateIcon обновляет существующую иконку
func (r *IconRepository) UpdateIcon(ctx context.Context, icon models.Icon) error {
	return r.db.WithContext(ctx).Model(&models.Icon{}).Where("id = ?", icon.ID).Updates(&icon).Error
}

// DeleteIcon удаляет иконку
func (r *IconRepository) DeleteIcon(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Icon{}).Error
}

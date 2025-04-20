package postgres

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// ActivityRepository представляет собой репозиторий для работы с активностями в PostgreSQL
type ActivityRepository struct {
	db *gorm.DB
}

// NewActivityRepository создает новый экземпляр ActivityRepository
func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{
		db: db,
	}
}

// GetActivitiesByEventID возвращает все активности по ID мероприятия
func (r *ActivityRepository) GetActivitiesByEventID(ctx context.Context, eventID int64) ([]models.Activity, error) {
	var activities []models.Activity
	if err := r.db.WithContext(ctx).Where("id_event = ?", eventID).Find(&activities).Error; err != nil {
		return nil, err
	}
	return activities, nil
}

// GetActivityByID возвращает активность по ID
func (r *ActivityRepository) GetActivityByID(ctx context.Context, id int) (models.Activity, error) {
	var activity models.Activity
	if err := r.db.WithContext(ctx).First(&activity, id).Error; err != nil {
		return models.Activity{}, err
	}
	return activity, nil
}

// CreateActivity создает новую активность
func (r *ActivityRepository) CreateActivity(ctx context.Context, activity models.Activity) (models.Activity, error) {
	if err := r.db.WithContext(ctx).Create(&activity).Error; err != nil {
		return models.Activity{}, err
	}
	return activity, nil
}

// UpdateActivity обновляет существующую активность
func (r *ActivityRepository) UpdateActivity(ctx context.Context, activity models.Activity) error {
	return r.db.WithContext(ctx).Save(&activity).Error
}

// DeleteActivity удаляет активность
func (r *ActivityRepository) DeleteActivity(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Activity{}, id).Error
}

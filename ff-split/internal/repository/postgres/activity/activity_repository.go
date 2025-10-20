package activity

import (
	"context"
	"errors"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// ActivityRepository реализует интерфейс repository.Activity
type ActivityRepository struct {
	db *gorm.DB
}

// NewActivityRepository создает новый экземпляр ActivityRepository
func NewActivityRepository(db *gorm.DB) *ActivityRepository {
	return &ActivityRepository{
		db: db,
	}
}

// GetByEventID возвращает активности по ID мероприятия
func (r *ActivityRepository) GetByEventID(ctx context.Context, eventID int64) ([]models.Activity, error) {
	var dbActivities []Activity
	err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&dbActivities).Error
	if err != nil {
		return nil, err
	}
	return extractSlice(dbActivities), nil
}

// GetByID возвращает активность по ID
func (r *ActivityRepository) GetByID(ctx context.Context, id int) (*models.Activity, error) {
	var dbActivity Activity
	err := r.db.WithContext(ctx).First(&dbActivity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // возвращаем nil, nil если активность не найдена
		}
		return nil, err
	}
	return extract(&dbActivity), nil
}

// Create создает новую активность
func (r *ActivityRepository) Create(ctx context.Context, activity *models.Activity) (*models.Activity, error) {
	dbActivity := load(activity)
	err := r.db.WithContext(ctx).Create(dbActivity).Error
	if err != nil {
		return nil, err
	}
	return extract(dbActivity), nil
}

// Update обновляет активность
func (r *ActivityRepository) Update(ctx context.Context, id int, activity *models.Activity) (*models.Activity, error) {
	// Проверяем существование активности
	var existingActivity Activity
	err := r.db.WithContext(ctx).First(&existingActivity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // возвращаем nil, nil если активность не найдена
		}
		return nil, err
	}

	// Обновляем только указанные поля
	dbActivity := load(activity)
	dbActivity.ID = id // Важно установить ID для правильного обновления
	err = r.db.WithContext(ctx).Model(&Activity{}).Where("id = ?", id).Updates(dbActivity).Error
	if err != nil {
		return nil, err
	}

	// Получаем обновленную активность
	var updatedActivity Activity
	err = r.db.WithContext(ctx).First(&updatedActivity, id).Error
	if err != nil {
		return nil, err
	}

	return extract(&updatedActivity), nil
}

// Delete удаляет активность
func (r *ActivityRepository) Delete(ctx context.Context, id int) error {
	result := r.db.WithContext(ctx).Delete(&Activity{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("активность не найдена")
	}

	return nil
}

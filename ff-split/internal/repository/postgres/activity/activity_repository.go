package activity

import (
	"context"
	"errors"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/history"
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

// Create сохраняет активность только в доступном для изменения событии.
func (r *ActivityRepository) Create(ctx context.Context, activity *models.Activity) (*models.Activity, error) {
	row := load(activity)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if row.EventID != nil {
			if err := history.LockEvent(tx, *row.EventID); err != nil {
				return err
			}
		}
		return tx.Create(row).Error
	})
	if err != nil {
		return nil, err
	}
	return extract(row), nil
}

// mutate блокирует сохранённый parent и запрещает перенос активности между событиями.
func (r *ActivityRepository) mutate(ctx context.Context, id int, activity *models.Activity) (*models.Activity, error) {
	var result *models.Activity
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row Activity
		if err := tx.First(&row, id).Error; err != nil {
			return err
		}
		if row.EventID != nil {
			if err := history.LockEvent(tx, *row.EventID); err != nil {
				return err
			}
		}
		if activity == nil {
			return tx.Delete(&Activity{}, id).Error
		}
		update := load(activity)
		if err := tx.Model(&Activity{}).Where("id=?", id).Select("description", "icon_id", "user_id").Updates(update).Error; err != nil {
			return err
		}
		if err := tx.First(&row, id).Error; err != nil {
			return err
		}
		result = extract(&row)
		return nil
	})
	return result, err
}

// Update меняет поля активности, сохраняя её исходное событие.
func (r *ActivityRepository) Update(ctx context.Context, id int, activity *models.Activity) (*models.Activity, error) {
	return r.mutate(ctx, id, activity)
}

// Delete удаляет активность только после проверки сохранённого события под блокировкой.
func (r *ActivityRepository) Delete(ctx context.Context, id int) error {
	_, err := r.mutate(ctx, id, nil)
	return err
}

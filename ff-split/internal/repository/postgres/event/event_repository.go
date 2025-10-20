package event

import (
	"context"
	"errors"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-split/internal/common/db"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// EventRepository реализует интерфейс repository.Event
type EventRepository struct {
	db *gorm.DB
}

// NewEventRepository создает новый экземпляр EventRepository
func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{
		db: db,
	}
}

// GetAll возвращает все мероприятия
func (r *EventRepository) GetAll(ctx context.Context) ([]models.Event, error) {
	var events []models.Event
	err := r.db.WithContext(ctx).Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

// GetByID возвращает мероприятие по ID
func (r *EventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	var event models.Event
	err := r.db.WithContext(ctx).First(&event, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // возвращаем nil, nil если мероприятие не найдено
		}
		return nil, err
	}
	return &event, nil
}

// Create создает новое мероприятие
func (r *EventRepository) Create(ctx context.Context, event *models.Event) error {
	return db.GetTx(ctx, r.db).WithContext(ctx).Create(event).Error
}

// Update обновляет мероприятие
func (r *EventRepository) Update(ctx context.Context, id int64, event *models.Event) error {
	conn := db.GetTx(ctx, r.db)

	// Проверка существования записи
	var exists bool
	if err := conn.Model(&models.Event{}).Select("count(*) > 0").Where("id = ?", id).Find(&exists).Error; err != nil {
		return fmt.Errorf("ошибка при проверке существования мероприятия: %w", err)
	}
	if !exists {
		return gorm.ErrRecordNotFound
	}

	event.ID = id
	if err := conn.Model(&models.Event{}).Where("id = ?", id).Updates(event).Error; err != nil {
		return fmt.Errorf("ошибка при обновлении мероприятия: %w", err)
	}

	return nil
}

// Delete удаляет мероприятие
func (r *EventRepository) Delete(ctx context.Context, id int64) error {
	err := db.WithTx(ctx, r.db, func(ctx context.Context) error {

		result := db.GetTx(ctx, r.db).WithContext(ctx).Delete(&models.UserEvent{}, "event_id = ?", id)
		if result.Error != nil {
			return result.Error
		}

		result = db.GetTx(ctx, r.db).WithContext(ctx).Delete(&models.Event{}, id)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("мероприятие не найдено")
		}

		return nil
	})
	return err
}


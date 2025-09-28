package postgres

import (
	"context"
	"errors"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// EventRepository реализует интерфейс repository.EventRepository
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
func (r *EventRepository) Create(ctx context.Context, event *models.Event) (*models.Event, error) {
	err := r.db.WithContext(ctx).Create(event).Error
	if err != nil {
		return nil, err
	}
	return event, nil
}

// Update обновляет мероприятие
func (r *EventRepository) Update(ctx context.Context, id int64, event *models.Event) (*models.Event, error) {
	// Проверяем существование мероприятия
	var existingEvent models.Event
	err := r.db.WithContext(ctx).First(&existingEvent, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // возвращаем nil, nil если мероприятие не найдено
		}
		return nil, err
	}

	// Обновляем только указанные поля
	event.ID = id // Важно установить ID для правильного обновления
	err = r.db.WithContext(ctx).Model(&models.Event{}).Where("id = ?", id).Updates(event).Error
	if err != nil {
		return nil, err
	}

	// Получаем обновленное мероприятие
	var updatedEvent models.Event
	err = r.db.WithContext(ctx).First(&updatedEvent, id).Error
	if err != nil {
		return nil, err
	}

	return &updatedEvent, nil
}

// Delete удаляет мероприятие
func (r *EventRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&models.Event{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("мероприятие не найдено")
	}

	return nil
}

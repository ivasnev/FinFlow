package postgres

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// EventRepository представляет собой репозиторий для работы с мероприятиями в PostgreSQL
type EventRepository struct {
	db *gorm.DB
}

// NewEventRepository создает новый экземпляр EventRepository
func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{
		db: db,
	}
}

// GetEvents возвращает все мероприятия
func (r *EventRepository) GetEvents(ctx context.Context) ([]models.Event, error) {
	var events []models.Event
	if err := r.db.WithContext(ctx).Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// GetEventByID возвращает мероприятие по ID
func (r *EventRepository) GetEventByID(ctx context.Context, id int64) (models.Event, error) {
	var event models.Event
	if err := r.db.WithContext(ctx).First(&event, id).Error; err != nil {
		return models.Event{}, err
	}
	return event, nil
}

// CreateEvent создает новое мероприятие
func (r *EventRepository) CreateEvent(ctx context.Context, event models.Event) (models.Event, error) {
	if err := r.db.WithContext(ctx).Create(&event).Error; err != nil {
		return models.Event{}, err
	}
	return event, nil
}

// UpdateEvent обновляет существующее мероприятие
func (r *EventRepository) UpdateEvent(ctx context.Context, event models.Event) error {
	return r.db.WithContext(ctx).Save(&event).Error
}

// DeleteEvent удаляет мероприятие
func (r *EventRepository) DeleteEvent(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Event{}, id).Error
}

// AddUserToEvent добавляет пользователя в мероприятие
func (r *EventRepository) AddUserToEvent(ctx context.Context, userEvent models.UserEvent) error {
	return r.db.WithContext(ctx).Create(&userEvent).Error
}

// RemoveUserFromEvent удаляет пользователя из мероприятия
func (r *EventRepository) RemoveUserFromEvent(ctx context.Context, userID, eventID int64) error {
	return r.db.WithContext(ctx).Where("id_user = ? AND id_event = ?", userID, eventID).Delete(&models.UserEvent{}).Error
}

// GetEventUsers возвращает всех пользователей мероприятия
func (r *EventRepository) GetEventUsers(ctx context.Context, eventID int64) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).
		Joins("JOIN user_event ON user_event.id_user = users.id_user").
		Where("user_event.id_event = ?", eventID).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// CreateDummyUser создает искусственного пользователя
func (r *EventRepository) CreateDummyUser(ctx context.Context, name string) (models.User, error) {
	user := models.User{
		NameCashed: name,
		IsDummy:    true,
	}

	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return models.User{}, err
	}

	return user, nil
}

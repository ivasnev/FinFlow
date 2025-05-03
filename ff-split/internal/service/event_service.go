package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
)

// EventService реализует интерфейс service.EventServiceInterface
type EventService struct {
	repo repository.EventRepository
}

// NewEventService создает новый экземпляр EventServiceInterface
func NewEventService(repo repository.EventRepository) *EventService {
	return &EventService{
		repo: repo,
	}
}

// GetEvents получает все мероприятия
func (s *EventService) GetEvents(ctx context.Context) ([]models.Event, error) {
	return s.repo.GetAll(ctx)
}

// GetEventByID получает мероприятие по ID
func (s *EventService) GetEventByID(ctx context.Context, id int64) (*models.Event, error) {
	return s.repo.GetByID(ctx, id)
}

// CreateEvent создает новое мероприятие
func (s *EventService) CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error) {
	return s.repo.Create(ctx, event)
}

// UpdateEvent обновляет мероприятие
func (s *EventService) UpdateEvent(ctx context.Context, id int64, event *models.Event) (*models.Event, error) {
	return s.repo.Update(ctx, id, event)
}

// DeleteEvent удаляет мероприятие
func (s *EventService) DeleteEvent(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

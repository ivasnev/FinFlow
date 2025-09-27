package service

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
)

// EventServiceImpl реализация сервиса мероприятий
type EventServiceImpl struct {
	eventRepo repository.EventRepository
}

// NewEventService создает новый экземпляр сервиса мероприятий
func NewEventService(eventRepo repository.EventRepository) *EventServiceImpl {
	return &EventServiceImpl{
		eventRepo: eventRepo,
	}
}

// GetEvents возвращает все мероприятия
func (s *EventServiceImpl) GetEvents(ctx context.Context) ([]dto.EventResponse, error) {
	events, err := s.eventRepo.GetEvents(ctx)
	if err != nil {
		return nil, err
	}

	// Преобразуем в DTO
	var response []dto.EventResponse
	for _, event := range events {
		response = append(response, mapEventToResponse(event))
	}

	return response, nil
}

// GetEventByID возвращает мероприятие по ID
func (s *EventServiceImpl) GetEventByID(ctx context.Context, id int64) (dto.EventResponse, error) {
	event, err := s.eventRepo.GetEventByID(ctx, id)
	if err != nil {
		return dto.EventResponse{}, err
	}

	return mapEventToResponse(event), nil
}

// CreateEvent создает новое мероприятие
func (s *EventServiceImpl) CreateEvent(ctx context.Context, eventRequest dto.EventRequest) (dto.EventResponse, error) {
	// Создаем модель мероприятия из запроса
	event := models.Event{
		Name:        eventRequest.Name,
		Description: eventRequest.Description,
		CategoryID:  eventRequest.CategoryID,
		Status:      "active", // По умолчанию мероприятие активно
	}

	// Создаем мероприятие
	createdEvent, err := s.eventRepo.CreateEvent(ctx, event)
	if err != nil {
		return dto.EventResponse{}, err
	}

	// Добавляем пользователей в мероприятие
	for _, userID := range eventRequest.Members.UserIDs {
		userEvent := models.UserEvent{
			IDUser:  userID,
			IDEvent: createdEvent.ID,
		}

		if err := s.eventRepo.AddUserToEvent(ctx, userEvent); err != nil {
			return dto.EventResponse{}, err
		}
	}

	// Создаем dummy пользователей, если указаны
	for _, dummyName := range eventRequest.Members.DummiesNames {
		dummyUser, err := s.eventRepo.CreateDummyUser(ctx, dummyName)
		if err != nil {
			return dto.EventResponse{}, err
		}

		userEvent := models.UserEvent{
			IDUser:  dummyUser.IDUser,
			IDEvent: createdEvent.ID,
		}

		if err := s.eventRepo.AddUserToEvent(ctx, userEvent); err != nil {
			return dto.EventResponse{}, err
		}
	}

	return mapEventToResponse(createdEvent), nil
}

// UpdateEvent обновляет существующее мероприятие
func (s *EventServiceImpl) UpdateEvent(ctx context.Context, id int64, eventRequest dto.EventRequest) error {
	// Проверяем наличие мероприятия
	event, err := s.eventRepo.GetEventByID(ctx, id)
	if err != nil {
		return err
	}

	// Обновляем данные
	event.Name = eventRequest.Name
	event.Description = eventRequest.Description
	event.CategoryID = eventRequest.CategoryID

	// Сохраняем изменения
	if err := s.eventRepo.UpdateEvent(ctx, event); err != nil {
		return err
	}

	// Удаляем всех текущих пользователей
	users, err := s.eventRepo.GetEventUsers(ctx, id)
	if err != nil {
		return err
	}

	for _, user := range users {
		if err := s.eventRepo.RemoveUserFromEvent(ctx, user.IDUser, id); err != nil {
			return err
		}
	}

	// Добавляем новых пользователей
	for _, userID := range eventRequest.Members.UserIDs {
		userEvent := models.UserEvent{
			IDUser:  userID,
			IDEvent: id,
		}

		if err := s.eventRepo.AddUserToEvent(ctx, userEvent); err != nil {
			return err
		}
	}

	// Создаем dummy пользователей, если указаны
	for _, dummyName := range eventRequest.Members.DummiesNames {
		dummyUser, err := s.eventRepo.CreateDummyUser(ctx, dummyName)
		if err != nil {
			return err
		}

		userEvent := models.UserEvent{
			IDUser:  dummyUser.IDUser,
			IDEvent: id,
		}

		if err := s.eventRepo.AddUserToEvent(ctx, userEvent); err != nil {
			return err
		}
	}

	return nil
}

// DeleteEvent удаляет мероприятие
func (s *EventServiceImpl) DeleteEvent(ctx context.Context, id int64) error {
	return s.eventRepo.DeleteEvent(ctx, id)
}

// mapEventToResponse преобразует модель Event в DTO
func mapEventToResponse(event models.Event) dto.EventResponse {
	return dto.EventResponse{
		ID:         event.ID,
		Name:       event.Name,
		CategoryID: event.CategoryID,
		PhotoID:    event.ImageID,
	}
}

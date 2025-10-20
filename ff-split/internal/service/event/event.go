package event

import (
	"context"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-split/internal/common/db"

	"gorm.io/gorm"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// EventService реализует интерфейс service.Event
type EventService struct {
	db          *gorm.DB
	userService service.User
	repo        repository.Event
}

// NewEventService создает новый экземпляр EventService
func NewEventService(repo repository.Event, dbImpl *gorm.DB, userService service.User) *EventService {
	return &EventService{
		repo:        repo,
		db:          dbImpl,
		userService: userService,
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
func (s *EventService) CreateEvent(ctx context.Context, request *service.EventRequest) (*service.EventResponse, error) {

	if request.Members.UserIDs != nil {
		notExistsUsers, err := s.userService.GetNotExistsUsers(ctx, request.Members.UserIDs)
		if err != nil {
			return nil, fmt.Errorf("Ошибка при получении несуществующих пользователей: %w", err)
		}
		if err := s.userService.BatchSyncUsersWithIDService(ctx, notExistsUsers); err != nil {
			return nil, fmt.Errorf("Ошибка при синхронизации пользователей: %w", err)
		}
	}

	// Преобразуем DTO в модель
	categoryID := request.CategoryID
	event := &models.Event{
		Name:        request.Name,
		Description: request.Description,
		CategoryID:  categoryID,
		Status:      "active", // Статус по умолчанию
	}

	err := db.WithTx(ctx, s.db, func(ctx context.Context) error {
		// Создаем мероприятие
		var err error
		err = s.repo.Create(ctx, event)
		if err != nil {
			return fmt.Errorf("Ошибка при создании мероприятия: %w", err)
		}
		var internalIds []int64
		if request.Members.DummiesNames != nil {
			dummies, createErr := s.userService.BatchCreateDummyUsers(ctx, request.Members.DummiesNames, event.ID)
			if createErr != nil {
				return fmt.Errorf("ошибка при создании dummy пользователей: %w", createErr)
			}
			for _, dummy := range dummies {
				internalIds = append(internalIds, dummy.ID)
			}
		}

		if request.Members.UserIDs != nil {
			userIds, err := s.userService.GetInternalUserIdsByExternalUserIds(ctx, request.Members.UserIDs)
			if err != nil {
				return fmt.Errorf("ошибка при получении внутренних ID пользователей: %w", err)
			}
			internalIds = append(internalIds, userIds...)
		}
		err = s.userService.AddUsersToEvent(ctx, internalIds, event.ID)
		if err != nil {
			return fmt.Errorf("ошибка при добавлении пользователей в мероприятие: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Заглушка для баланса
	var balance *int = nil
	// Здесь будет расчет баланса в будущем

	return &service.EventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		CategoryID:  event.CategoryID,
		PhotoID:     event.ImageID,
		Balance:     balance,
	}, nil
}

// UpdateEvent обновляет мероприятие
func (s *EventService) UpdateEvent(ctx context.Context, id int64, request *service.EventRequest) (*service.EventResponse, error) {
	// Преобразуем DTO в модель
	event := &models.Event{
		Name:        request.Name,
		Description: request.Description,
		CategoryID:  request.CategoryID,
	}

	err := db.WithTx(ctx, s.db, func(ctx context.Context) error {
		err := s.repo.Update(ctx, id, event)
		if err != nil {
			return fmt.Errorf("Ошибка при обновлении мероприятия: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Заглушка для баланса
	var balance *int = nil
	// Здесь будет расчет баланса в будущем

	return &service.EventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		CategoryID:  event.CategoryID,
		PhotoID:     event.ImageID,
		Balance:     balance,
	}, nil
}

// DeleteEvent удаляет мероприятие
func (s *EventService) DeleteEvent(ctx context.Context, id int64) error {
	return db.WithTx(ctx, s.db, func(ctx context.Context) error {
		err := s.repo.Delete(ctx, id)
		if err != nil {
			return fmt.Errorf("Ошибка при удалении мероприятия: %w", err)
		}
		return nil
	})
}

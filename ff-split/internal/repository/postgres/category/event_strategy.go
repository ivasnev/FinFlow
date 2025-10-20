package category

import (
	"context"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	"gorm.io/gorm"
)

// EventCategoryStrategy реализует стратегию для работы с категориями мероприятий
type EventCategoryStrategy struct {
	db *gorm.DB
}

// NewEventCategoryStrategy создает новую стратегию для работы с категориями мероприятий
func NewEventCategoryStrategy(db *gorm.DB) *EventCategoryStrategy {
	return &EventCategoryStrategy{
		db: db,
	}
}

// GetAll возвращает все категории мероприятий
func (s *EventCategoryStrategy) GetAll(ctx context.Context) ([]service.CategoryDTO, error) {
	var eventCategories []EventCategory
	if err := s.db.WithContext(ctx).Find(&eventCategories).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении категорий мероприятий: %w", err)
	}

	result := make([]service.CategoryDTO, len(eventCategories))
	for i, category := range eventCategories {
		result[i] = service.CategoryDTO{
			ID:     category.ID,
			Name:   category.Name,
			IconID: category.IconID,
			Icon: service.IconDTO{
				ID:           category.IconID,
				Name:         "",
				ExternalUuid: "",
			},
		}
	}

	return result, nil
}

// GetByID возвращает категорию мероприятия по ID
func (s *EventCategoryStrategy) GetByID(ctx context.Context, id int) (*service.CategoryDTO, error) {
	var eventCategory EventCategory
	if err := s.db.WithContext(ctx).First(&eventCategory, id).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении категории мероприятия: %w", err)
	}
	return &service.CategoryDTO{
		ID:     eventCategory.ID,
		Name:   eventCategory.Name,
		IconID: eventCategory.IconID,
		Icon: service.IconDTO{
			ID:           eventCategory.IconID,
			Name:         "",
			ExternalUuid: "",
		},
	}, nil
}

// Create создает новую категорию мероприятия
func (s *EventCategoryStrategy) Create(ctx context.Context, category *service.CategoryDTO) (*service.CategoryDTO, error) {
	eventCategory := EventCategory{
		Name:   category.Name,
		IconID: category.IconID,
	}

	if err := s.db.WithContext(ctx).Create(&eventCategory).Error; err != nil {
		return nil, fmt.Errorf("ошибка при создании категории мероприятия: %w", err)
	}
	return &service.CategoryDTO{
		ID:     eventCategory.ID,
		Name:   eventCategory.Name,
		IconID: eventCategory.IconID,
		Icon: service.IconDTO{
			ID:           eventCategory.IconID,
			Name:         "",
			ExternalUuid: "",
		},
	}, nil
}

// Update обновляет существующую категорию мероприятия
func (s *EventCategoryStrategy) Update(ctx context.Context, category *service.CategoryDTO) (*service.CategoryDTO, error) {
	eventCategory := EventCategory{
		ID:     category.ID,
		Name:   category.Name,
		IconID: category.IconID,
	}

	result := s.db.WithContext(ctx).Save(&eventCategory)
	if result.Error != nil {
		return nil, fmt.Errorf("ошибка при обновлении категории мероприятия: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("категория мероприятия с ID %d не найдена", category.ID)
	}

	return &service.CategoryDTO{
		ID:     eventCategory.ID,
		Name:   eventCategory.Name,
		IconID: eventCategory.IconID,
		Icon: service.IconDTO{
			ID:           eventCategory.IconID,
			Name:         "",
			ExternalUuid: "",
		},
	}, nil
}

// Delete удаляет категорию мероприятия
func (s *EventCategoryStrategy) Delete(ctx context.Context, id int) error {
	result := s.db.WithContext(ctx).Delete(&EventCategory{}, id)
	if result.Error != nil {
		return fmt.Errorf("ошибка при удалении категории мероприятия: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("категория мероприятия с ID %d не найдена", id)
	}

	return nil
}

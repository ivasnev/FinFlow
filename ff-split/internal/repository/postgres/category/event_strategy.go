package category

import (
	"context"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
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
func (s *EventCategoryStrategy) GetAll(ctx context.Context) ([]dto.CategoryDTO, error) {
	var eventCategories []models.EventCategory
	if err := s.db.WithContext(ctx).Preload("Icon").Find(&eventCategories).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении категорий мероприятий: %w", err)
	}

	result := make([]dto.CategoryDTO, len(eventCategories))
	for i, category := range eventCategories {
		categoryDto := &dto.CategoryDTO{
			ID:     category.ID,
			Name:   category.Name,
			IconID: category.IconID,
		}
		if category.Icon != nil {
			categoryDto.Icon = dto.IconDTO{
				ID:           category.Icon.ID,
				Name:         category.Icon.Name,
				ExternalUuid: category.Icon.FileUUID,
			}
		}
		result[i] = *categoryDto
	}

	return result, nil
}

// GetByID возвращает категорию мероприятия по ID
func (s *EventCategoryStrategy) GetByID(ctx context.Context, id int) (*dto.CategoryDTO, error) {
	var eventCategory models.EventCategory
	if err := s.db.WithContext(ctx).Preload("Icon").First(&eventCategory, id).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении категории мероприятия: %w", err)
	}
	categoryDto := &dto.CategoryDTO{
		ID:     eventCategory.ID,
		Name:   eventCategory.Name,
		IconID: eventCategory.IconID,
	}
	if eventCategory.Icon != nil {
		categoryDto.Icon = dto.IconDTO{
			ID:           eventCategory.Icon.ID,
			Name:         eventCategory.Icon.Name,
			ExternalUuid: eventCategory.Icon.FileUUID,
		}
	}
	return categoryDto, nil
}

// Create создает новую категорию мероприятия
func (s *EventCategoryStrategy) Create(ctx context.Context, category *dto.CategoryDTO) (*dto.CategoryDTO, error) {
	eventCategory := models.EventCategory{
		Name:   category.Name,
		IconID: category.IconID,
	}

	if err := s.db.WithContext(ctx).Create(&eventCategory).Error; err != nil {
		return nil, fmt.Errorf("ошибка при создании категории мероприятия: %w", err)
	}
	categoryDto := &dto.CategoryDTO{
		ID:     eventCategory.ID,
		Name:   eventCategory.Name,
		IconID: eventCategory.IconID,
	}
	if eventCategory.Icon != nil {
		categoryDto.Icon = dto.IconDTO{
			ID:           eventCategory.Icon.ID,
			Name:         eventCategory.Icon.Name,
			ExternalUuid: eventCategory.Icon.FileUUID,
		}
	}
	return categoryDto, nil
}

// Update обновляет существующую категорию мероприятия
func (s *EventCategoryStrategy) Update(ctx context.Context, category *dto.CategoryDTO) (*dto.CategoryDTO, error) {
	eventCategory := models.EventCategory{
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

	categoryDto := &dto.CategoryDTO{
		ID:     eventCategory.ID,
		Name:   eventCategory.Name,
		IconID: eventCategory.IconID,
	}
	if eventCategory.Icon != nil {
		categoryDto.Icon = dto.IconDTO{
			ID:           eventCategory.Icon.ID,
			Name:         eventCategory.Icon.Name,
			ExternalUuid: eventCategory.Icon.FileUUID,
		}
	}
	return categoryDto, nil
}

// Delete удаляет категорию мероприятия
func (s *EventCategoryStrategy) Delete(ctx context.Context, id int) error {
	result := s.db.WithContext(ctx).Delete(&models.EventCategory{}, id)
	if result.Error != nil {
		return fmt.Errorf("ошибка при удалении категории мероприятия: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("категория мероприятия с ID %d не найдена", id)
	}

	return nil
}

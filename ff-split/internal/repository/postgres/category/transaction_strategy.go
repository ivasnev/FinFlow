package category

import (
	"context"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// TransactionCategoryStrategy реализует стратегию для работы с категориями транзакций
type TransactionCategoryStrategy struct {
	db *gorm.DB
}

// NewTransactionCategoryStrategy создает новую стратегию для работы с категориями транзакций
func NewTransactionCategoryStrategy(db *gorm.DB) *TransactionCategoryStrategy {
	return &TransactionCategoryStrategy{
		db: db,
	}
}

// GetAll возвращает все категории транзакций
func (s *TransactionCategoryStrategy) GetAll(ctx context.Context) ([]dto.CategoryDTO, error) {
	var transactionCategories []models.TransactionCategory
	if err := s.db.WithContext(ctx).Preload("Icon").Find(&transactionCategories).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении категорий транзакций: %w", err)
	}

	result := make([]dto.CategoryDTO, len(transactionCategories))
	for i, category := range transactionCategories {
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

// GetByID возвращает категорию транзакции по ID
func (s *TransactionCategoryStrategy) GetByID(ctx context.Context, id int) (*dto.CategoryDTO, error) {
	var transactionCategory models.TransactionCategory
	if err := s.db.WithContext(ctx).Preload("Icon").First(&transactionCategory, id).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении категории транзакции: %w", err)
	}
	categoryDto := &dto.CategoryDTO{
		ID:     transactionCategory.ID,
		Name:   transactionCategory.Name,
		IconID: transactionCategory.IconID,
	}
	if transactionCategory.Icon != nil {
		categoryDto.Icon = dto.IconDTO{
			ID:           transactionCategory.Icon.ID,
			Name:         transactionCategory.Icon.Name,
			ExternalUuid: transactionCategory.Icon.FileUUID,
		}
	}

	return categoryDto, nil
}

// Create создает новую категорию транзакции
func (s *TransactionCategoryStrategy) Create(ctx context.Context, category *dto.CategoryDTO) (*dto.CategoryDTO, error) {
	transactionCategory := models.TransactionCategory{
		Name:   category.Name,
		IconID: category.IconID,
	}

	if err := s.db.WithContext(ctx).Create(&transactionCategory).Error; err != nil {
		return nil, fmt.Errorf("ошибка при создании категории транзакции: %w", err)
	}
	categoryDto := &dto.CategoryDTO{
		ID:     transactionCategory.ID,
		Name:   transactionCategory.Name,
		IconID: transactionCategory.IconID,
	}
	if transactionCategory.Icon != nil {
		categoryDto.Icon = dto.IconDTO{
			ID:           transactionCategory.Icon.ID,
			Name:         transactionCategory.Icon.Name,
			ExternalUuid: transactionCategory.Icon.FileUUID,
		}
	}

	return categoryDto, nil
}

// Update обновляет существующую категорию транзакции
func (s *TransactionCategoryStrategy) Update(ctx context.Context, category *dto.CategoryDTO) (*dto.CategoryDTO, error) {
	transactionCategory := models.TransactionCategory{
		ID:     category.ID,
		Name:   category.Name,
		IconID: category.IconID,
	}

	result := s.db.WithContext(ctx).Save(&transactionCategory)
	if result.Error != nil {
		return nil, fmt.Errorf("ошибка при обновлении категории транзакции: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("категория транзакции с ID %d не найдена", category.ID)
	}

	categoryDto := &dto.CategoryDTO{
		ID:     transactionCategory.ID,
		Name:   transactionCategory.Name,
		IconID: transactionCategory.IconID,
	}
	if transactionCategory.Icon != nil {
		categoryDto.Icon = dto.IconDTO{
			ID:           transactionCategory.Icon.ID,
			Name:         transactionCategory.Icon.Name,
			ExternalUuid: transactionCategory.Icon.FileUUID,
		}
	}

	return categoryDto, nil
}

// Delete удаляет категорию транзакции
func (s *TransactionCategoryStrategy) Delete(ctx context.Context, id int) error {
	result := s.db.WithContext(ctx).Delete(&models.TransactionCategory{}, id)
	if result.Error != nil {
		return fmt.Errorf("ошибка при удалении категории транзакции: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("категория транзакции с ID %d не найдена", id)
	}

	return nil
}

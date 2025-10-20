package category

import (
	"context"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-split/internal/service"
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
func (s *TransactionCategoryStrategy) GetAll(ctx context.Context) ([]service.CategoryDTO, error) {
	var transactionCategories []TransactionCategory
	if err := s.db.WithContext(ctx).Find(&transactionCategories).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении категорий транзакций: %w", err)
	}

	result := make([]service.CategoryDTO, len(transactionCategories))
	for i, category := range transactionCategories {
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

// GetByID возвращает категорию транзакции по ID
func (s *TransactionCategoryStrategy) GetByID(ctx context.Context, id int) (*service.CategoryDTO, error) {
	var transactionCategory TransactionCategory
	if err := s.db.WithContext(ctx).First(&transactionCategory, id).Error; err != nil {
		return nil, fmt.Errorf("ошибка при получении категории транзакции: %w", err)
	}
	return &service.CategoryDTO{
		ID:     transactionCategory.ID,
		Name:   transactionCategory.Name,
		IconID: transactionCategory.IconID,
		Icon: service.IconDTO{
			ID:           transactionCategory.IconID,
			Name:         "",
			ExternalUuid: "",
		},
	}, nil
}

// Create создает новую категорию транзакции
func (s *TransactionCategoryStrategy) Create(ctx context.Context, category *service.CategoryDTO) (*service.CategoryDTO, error) {
	transactionCategory := TransactionCategory{
		Name:   category.Name,
		IconID: category.IconID,
	}

	if err := s.db.WithContext(ctx).Create(&transactionCategory).Error; err != nil {
		return nil, fmt.Errorf("ошибка при создании категории транзакции: %w", err)
	}
	return &service.CategoryDTO{
		ID:     transactionCategory.ID,
		Name:   transactionCategory.Name,
		IconID: transactionCategory.IconID,
		Icon: service.IconDTO{
			ID:           transactionCategory.IconID,
			Name:         "",
			ExternalUuid: "",
		},
	}, nil
}

// Update обновляет существующую категорию транзакции
func (s *TransactionCategoryStrategy) Update(ctx context.Context, category *service.CategoryDTO) (*service.CategoryDTO, error) {
	transactionCategory := TransactionCategory{
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

	return &service.CategoryDTO{
		ID:     transactionCategory.ID,
		Name:   transactionCategory.Name,
		IconID: transactionCategory.IconID,
		Icon: service.IconDTO{
			ID:           transactionCategory.IconID,
			Name:         "",
			ExternalUuid: "",
		},
	}, nil
}

// Delete удаляет категорию транзакции
func (s *TransactionCategoryStrategy) Delete(ctx context.Context, id int) error {
	result := s.db.WithContext(ctx).Delete(&TransactionCategory{}, id)
	if result.Error != nil {
		return fmt.Errorf("ошибка при удалении категории транзакции: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("категория транзакции с ID %d не найдена", id)
	}

	return nil
}

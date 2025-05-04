package category

import (
	"context"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"gorm.io/gorm"
)

type CategoryStrategy interface {
	GetAll(ctx context.Context) ([]dto.CategoryDTO, error)
	GetByID(ctx context.Context, id int) (*dto.CategoryDTO, error)
	Create(ctx context.Context, dto *dto.CategoryDTO) (*dto.CategoryDTO, error)
	Update(ctx context.Context, dto *dto.CategoryDTO) (*dto.CategoryDTO, error)
	Delete(ctx context.Context, id int) error
}

// Repository представляет репозиторий для работы с категориями
type Repository struct {
	db         *gorm.DB
	strategies map[string]CategoryStrategy
}

// NewRepository создает новый репозиторий для работы с категориями
func NewRepository(db *gorm.DB) *Repository {
	repo := &Repository{
		db:         db,
		strategies: make(map[string]CategoryStrategy),
	}

	// Регистрация стратегий для разных типов категорий
	repo.strategies["event"] = NewEventCategoryStrategy(repo.db)
	repo.strategies["transaction"] = NewTransactionCategoryStrategy(repo.db)

	return repo
}

func (r *Repository) GetCategoryTypes() ([]string, error) {
	types := make([]string, 0, len(r.strategies))
	for t := range r.strategies {
		types = append(types, t)
	}
	return types, nil
}

// GetAll получает все категории указанного типа
func (r *Repository) GetAll(ctx context.Context, categoryType string) ([]dto.CategoryDTO, error) {
	strategy, ok := r.strategies[categoryType]
	if !ok {
		return nil, fmt.Errorf("неизвестный тип категории: %s", categoryType)
	}
	return strategy.GetAll(ctx)
}

// GetByID получает категорию по ID и типу
func (r *Repository) GetByID(ctx context.Context, categoryType string, id int) (*dto.CategoryDTO, error) {
	strategy, ok := r.strategies[categoryType]
	if !ok {
		return nil, fmt.Errorf("неизвестный тип категории: %s", categoryType)
	}
	return strategy.GetByID(ctx, id)
}

// Create создает новую категорию
func (r *Repository) Create(ctx context.Context, categoryType string, dto *dto.CategoryDTO) (*dto.CategoryDTO, error) {
	strategy, ok := r.strategies[categoryType]
	if !ok {
		return nil, fmt.Errorf("неизвестный тип категории: %s", categoryType)
	}
	return strategy.Create(ctx, dto)
}

// Update обновляет существующую категорию
func (r *Repository) Update(ctx context.Context, categoryType string, dto *dto.CategoryDTO) (*dto.CategoryDTO, error) {
	strategy, ok := r.strategies[categoryType]
	if !ok {
		return nil, fmt.Errorf("неизвестный тип категории: %s", categoryType)
	}
	return strategy.Update(ctx, dto)
}

// Delete удаляет категорию
func (r *Repository) Delete(ctx context.Context, categoryType string, id int) error {
	strategy, ok := r.strategies[categoryType]
	if !ok {
		return fmt.Errorf("неизвестный тип категории: %s", categoryType)
	}
	return strategy.Delete(ctx, id)
}

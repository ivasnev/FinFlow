package category

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
)

// CategoryService реализует интерфейс service.Category
type CategoryService struct {
	repo repository.Category
}

// NewCategoryService создает новый экземпляр CategoryService
func NewCategoryService(repo repository.Category) *CategoryService {
	return &CategoryService{
		repo: repo,
	}
}

// GetCategories получает все категории указанного типа
func (s *CategoryService) GetCategories(ctx context.Context, categoryType string) ([]dto.CategoryDTO, error) {
	return s.repo.GetAll(ctx, categoryType)
}

// GetCategoryByID получает категорию по ID и типу
func (s *CategoryService) GetCategoryByID(ctx context.Context, id int, categoryType string) (*dto.CategoryDTO, error) {
	return s.repo.GetByID(ctx, categoryType, id)
}

// CreateCategory создает новую категорию указанного типа
func (s *CategoryService) CreateCategory(ctx context.Context, category *dto.CategoryDTO, categoryType string) (*dto.CategoryDTO, error) {
	return s.repo.Create(ctx, categoryType, category)
}

// UpdateCategory обновляет категорию указанного типа
func (s *CategoryService) UpdateCategory(ctx context.Context, id int, category *dto.CategoryDTO, categoryType string) (*dto.CategoryDTO, error) {
	return s.repo.Update(ctx, categoryType, category)
}

// DeleteCategory удаляет категорию указанного типа
func (s *CategoryService) DeleteCategory(ctx context.Context, id int, categoryType string) error {
	return s.repo.Delete(ctx, categoryType, id)
}

// GetCategoryTypes возвращает список доступных типов категорий
func (s *CategoryService) GetCategoryTypes() ([]string, error) {
	return s.repo.GetCategoryTypes()
}

package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
)

// CategoryService реализует интерфейс service.CategoryServiceInterface
type CategoryService struct {
	repo repository.CategoryRepository
}

// NewCategoryService создает новый экземпляр CategoryServiceInterface
func NewCategoryService(repo repository.CategoryRepository) *CategoryService {
	return &CategoryService{
		repo: repo,
	}
}

// GetCategories получает все категории указанного типа
func (s *CategoryService) GetCategories(ctx context.Context, categoryType string) ([]models.EventCategory, error) {
	return s.repo.GetAll(ctx, categoryType)
}

// GetCategoryByID получает категорию по ID и типу
func (s *CategoryService) GetCategoryByID(ctx context.Context, id int, categoryType string) (*models.EventCategory, error) {
	return s.repo.GetByID(ctx, id, categoryType)
}

// CreateCategory создает новую категорию указанного типа
func (s *CategoryService) CreateCategory(ctx context.Context, category *models.EventCategory, categoryType string) (*models.EventCategory, error) {
	return s.repo.Create(ctx, category, categoryType)
}

// UpdateCategory обновляет категорию указанного типа
func (s *CategoryService) UpdateCategory(ctx context.Context, id int, category *models.EventCategory, categoryType string) (*models.EventCategory, error) {
	return s.repo.Update(ctx, id, category, categoryType)
}

// DeleteCategory удаляет категорию указанного типа
func (s *CategoryService) DeleteCategory(ctx context.Context, id int, categoryType string) error {
	return s.repo.Delete(ctx, id, categoryType)
}

// GetCategoryTypes возвращает список доступных типов категорий
func (s *CategoryService) GetCategoryTypes(ctx context.Context) ([]string, error) {
	return s.repo.GetCategoryTypes(ctx)
}

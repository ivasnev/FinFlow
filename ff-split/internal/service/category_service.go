package service

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
)

// CategoryServiceImpl реализация сервиса категорий
type CategoryServiceImpl struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryService создает новый экземпляр сервиса категорий
func NewCategoryService(categoryRepo repository.CategoryRepository) *CategoryServiceImpl {
	return &CategoryServiceImpl{
		categoryRepo: categoryRepo,
	}
}

// GetCategories возвращает все категории
func (s *CategoryServiceImpl) GetCategories(ctx context.Context) ([]dto.CategoryResponse, error) {
	categories, err := s.categoryRepo.GetCategories(ctx)
	if err != nil {
		return nil, err
	}

	var response []dto.CategoryResponse
	for _, category := range categories {
		response = append(response, mapCategoryToResponse(category))
	}

	return response, nil
}

// GetCategoryByID возвращает категорию по ID
func (s *CategoryServiceImpl) GetCategoryByID(ctx context.Context, id int64) (dto.CategoryResponse, error) {
	category, err := s.categoryRepo.GetCategoryByID(ctx, id)
	if err != nil {
		return dto.CategoryResponse{}, err
	}

	return mapCategoryToResponse(category), nil
}

// CreateCategory создает новую категорию
func (s *CategoryServiceImpl) CreateCategory(ctx context.Context, name string, iconID int64) (dto.CategoryResponse, error) {
	category := models.Category{
		Name:   name,
		IconID: iconID,
	}

	createdCategory, err := s.categoryRepo.CreateCategory(ctx, category)
	if err != nil {
		return dto.CategoryResponse{}, err
	}

	return mapCategoryToResponse(createdCategory), nil
}

// UpdateCategory обновляет существующую категорию
func (s *CategoryServiceImpl) UpdateCategory(ctx context.Context, id int64, name string, iconID int64) error {
	category := models.Category{
		ID:     id,
		Name:   name,
		IconID: iconID,
	}

	return s.categoryRepo.UpdateCategory(ctx, category)
}

// DeleteCategory удаляет категорию
func (s *CategoryServiceImpl) DeleteCategory(ctx context.Context, id int64) error {
	return s.categoryRepo.DeleteCategory(ctx, id)
}

// mapCategoryToResponse преобразует модель Category в DTO
func mapCategoryToResponse(category models.Category) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:     category.ID,
		Name:   category.Name,
		IconID: category.IconID,
	}
}

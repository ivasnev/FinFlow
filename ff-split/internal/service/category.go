package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
)

// Category определяет методы для работы с категориями
type Category interface {
	GetCategories(ctx context.Context, categoryType string) ([]dto.CategoryDTO, error)
	GetCategoryByID(ctx context.Context, id int, categoryType string) (*dto.CategoryDTO, error)
	CreateCategory(ctx context.Context, category *dto.CategoryDTO, categoryType string) (*dto.CategoryDTO, error)
	UpdateCategory(ctx context.Context, id int, category *dto.CategoryDTO, categoryType string) (*dto.CategoryDTO, error)
	DeleteCategory(ctx context.Context, id int, categoryType string) error
	GetCategoryTypes() ([]string, error)
}

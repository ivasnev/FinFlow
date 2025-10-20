package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
)

// Category определяет методы для работы с категориями
type Category interface {
	GetAll(ctx context.Context, categoryType string) ([]dto.CategoryDTO, error)
	GetByID(ctx context.Context, categoryType string, id int) (*dto.CategoryDTO, error)
	Create(ctx context.Context, categoryType string, category *dto.CategoryDTO) (*dto.CategoryDTO, error)
	Update(ctx context.Context, categoryType string, category *dto.CategoryDTO) (*dto.CategoryDTO, error)
	Delete(ctx context.Context, categoryType string, id int) error
	GetCategoryTypes() ([]string, error)
}


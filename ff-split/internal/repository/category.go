package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// Category определяет методы для работы с категориями
type Category interface {
	GetAll(ctx context.Context, categoryType string) ([]service.CategoryDTO, error)
	GetByID(ctx context.Context, categoryType string, id int) (*service.CategoryDTO, error)
	Create(ctx context.Context, categoryType string, category *service.CategoryDTO) (*service.CategoryDTO, error)
	Update(ctx context.Context, categoryType string, category *service.CategoryDTO) (*service.CategoryDTO, error)
	Delete(ctx context.Context, categoryType string, id int) error
	GetCategoryTypes() ([]string, error)
}

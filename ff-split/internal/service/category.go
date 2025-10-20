package service

import (
	"context"
)

// CategoryRequest представляет DTO для запроса создания/обновления категории
type CategoryRequest struct {
	Name   string `json:"name" binding:"required"`
	IconID int    `json:"icon_id"`
}

// CategoryDTO представляет DTO для передачи данных категорий
type CategoryDTO struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	IconID int    `json:"icon_id"`

	Icon IconDTO `json:"icon"`
}

// IconDTO представляет DTO для иконки в категории
type IconDTO struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ExternalUuid string `json:"external_uuid"`
}

// CategoryResponse представляет DTO для ответа с данными категории
type CategoryResponse struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`
	Icon IconDTO `json:"icon"`
}

// CategoryListResponse представляет DTO для ответа со списком категорий
type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

// CategoryTypesResponse представляет DTO для ответа со списком типов категорий
type CategoryTypesResponse []string

// Category определяет методы для работы с категориями
type Category interface {
	GetCategories(ctx context.Context, categoryType string) ([]CategoryDTO, error)
	GetCategoryByID(ctx context.Context, id int, categoryType string) (*CategoryDTO, error)
	CreateCategory(ctx context.Context, category *CategoryDTO, categoryType string) (*CategoryDTO, error)
	UpdateCategory(ctx context.Context, id int, category *CategoryDTO, categoryType string) (*CategoryDTO, error)
	DeleteCategory(ctx context.Context, id int, categoryType string) error
	GetCategoryTypes() ([]string, error)
}

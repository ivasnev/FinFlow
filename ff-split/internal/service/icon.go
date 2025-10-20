package service

import (
	"context"
)

// IconFullDTO представляет полное DTO для иконки
type IconFullDTO struct {
	ID       uint   `json:"id"`
	Name     string `json:"name" binding:"required"`
	FileUUID string `json:"file_uuid" binding:"required"`
}

// IconResponse представляет ответ на операцию с иконкой
type IconResponse IconFullDTO

// IconListResponse представляет ответ со списком иконок
type IconListResponse []IconFullDTO

// Icon определяет методы для работы с иконками
type Icon interface {
	GetIcons(ctx context.Context) ([]IconFullDTO, error)
	GetIconByID(ctx context.Context, id uint) (*IconFullDTO, error)
	CreateIcon(ctx context.Context, icon *IconFullDTO) (*IconFullDTO, error)
	UpdateIcon(ctx context.Context, id uint, icon *IconFullDTO) (*IconFullDTO, error)
	DeleteIcon(ctx context.Context, id uint) error
}

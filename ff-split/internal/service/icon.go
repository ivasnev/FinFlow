package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
)

// Icon определяет методы для работы с иконками
type Icon interface {
	GetIcons(ctx context.Context) ([]dto.IconFullDTO, error)
	GetIconByID(ctx context.Context, id uint) (*dto.IconFullDTO, error)
	CreateIcon(ctx context.Context, icon *dto.IconFullDTO) (*dto.IconFullDTO, error)
	UpdateIcon(ctx context.Context, id uint, icon *dto.IconFullDTO) (*dto.IconFullDTO, error)
	DeleteIcon(ctx context.Context, id uint) error
}


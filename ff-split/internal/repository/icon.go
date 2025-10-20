package repository

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Icon определяет методы для работы с иконками
type Icon interface {
	GetIcons() ([]models.Icon, error)
	GetIconByID(id uint) (*models.Icon, error)
	CreateIcon(icon *models.Icon) error
	UpdateIcon(icon *models.Icon) error
	DeleteIcon(id uint) error
}

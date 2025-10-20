package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Activity определяет методы для работы с активностями
type Activity interface {
	GetByEventID(ctx context.Context, eventID int64) ([]models.Activity, error)
	GetByID(ctx context.Context, id int) (*models.Activity, error)
	Create(ctx context.Context, activity *models.Activity) (*models.Activity, error)
	Update(ctx context.Context, id int, activity *models.Activity) (*models.Activity, error)
	Delete(ctx context.Context, id int) error
}


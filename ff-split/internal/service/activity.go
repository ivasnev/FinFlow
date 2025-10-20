package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Activity определяет методы для работы с активностями
type Activity interface {
	GetActivitiesByEventID(ctx context.Context, eventID int64) ([]models.Activity, error)
	GetActivityByID(ctx context.Context, id int) (*models.Activity, error)
	CreateActivity(ctx context.Context, activity *models.Activity) (*models.Activity, error)
	UpdateActivity(ctx context.Context, id int, activity *models.Activity) (*models.Activity, error)
	DeleteActivity(ctx context.Context, id int) error
}


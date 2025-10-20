package service

import (
	"context"
	"time"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// ActivityRequest представляет DTO для запроса создания/обновления активности
type ActivityRequest struct {
	UserID      *int64 `json:"user_id"`
	Description string `json:"description" binding:"required"`
	IconID      int    `json:"icon_id"`
}

// ActivityResponse представляет DTO для ответа с данными активности
type ActivityResponse struct {
	ActivityID  int       `json:"activity_id"`
	Description string    `json:"description"`
	IconID      int       `json:"icon_id"`
	Datetime    time.Time `json:"datetime"`
}

// ActivityListResponse представляет DTO для ответа со списком активностей
type ActivityListResponse struct {
	Activities []ActivityResponse `json:"activities"`
}

// Activity определяет методы для работы с активностями
type Activity interface {
	GetActivitiesByEventID(ctx context.Context, eventID int64) ([]models.Activity, error)
	GetActivityByID(ctx context.Context, id int) (*models.Activity, error)
	CreateActivity(ctx context.Context, activity *models.Activity) (*models.Activity, error)
	UpdateActivity(ctx context.Context, id int, activity *models.Activity) (*models.Activity, error)
	DeleteActivity(ctx context.Context, id int) error
}

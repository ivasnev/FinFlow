package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Event определяет методы для работы с мероприятиями
type Event interface {
	GetAll(ctx context.Context) ([]models.Event, error)
	GetByID(ctx context.Context, id int64) (*models.Event, error)
	GetByUserID(ctx context.Context, userID int64) ([]models.Event, error)
	CalculateUserBalances(ctx context.Context, userID int64, eventIDs []int64) (map[int64]float64, error)
	Create(ctx context.Context, event *models.Event) error
	Update(ctx context.Context, id int64, event *models.Event) error
	Delete(ctx context.Context, id int64) error
}

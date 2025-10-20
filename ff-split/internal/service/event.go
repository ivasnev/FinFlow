package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Event определяет методы для работы с мероприятиями
type Event interface {
	GetEvents(ctx context.Context) ([]models.Event, error)
	GetEventByID(ctx context.Context, id int64) (*models.Event, error)
	CreateEvent(ctx context.Context, request *dto.EventRequest) (*dto.EventResponse, error)
	UpdateEvent(ctx context.Context, id int64, request *dto.EventRequest) (*dto.EventResponse, error)
	DeleteEvent(ctx context.Context, id int64) error
}


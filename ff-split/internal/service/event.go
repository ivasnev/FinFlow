package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// EventRequest представляет DTO для запроса создания/обновления мероприятия
type EventRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	CategoryID  *int            `json:"category_id,omitempty"`
	Members     EventMembersDTO `json:"members"`
}

// EventMembersDTO представляет DTO для передачи данных о членах мероприятия
type EventMembersDTO struct {
	UserIDs      []int64  `json:"user_ids"`
	DummiesNames []string `json:"dummies_names"`
}

// EventResponse представляет DTO для ответа с данными мероприятия
type EventResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CategoryID  *int   `json:"category_id,omitempty"`
	PhotoID     string `json:"photo_id,omitempty"`
	Balance     *int   `json:"balance,omitempty"`
}

// EventListResponse представляет DTO для ответа со списком мероприятий
type EventListResponse struct {
	Events []EventResponse `json:"events"`
}

// Event определяет методы для работы с мероприятиями
type Event interface {
	GetEvents(ctx context.Context) ([]models.Event, error)
	GetEventByID(ctx context.Context, id int64) (*models.Event, error)
	CreateEvent(ctx context.Context, request *EventRequest) (*EventResponse, error)
	UpdateEvent(ctx context.Context, id int64, request *EventRequest) (*EventResponse, error)
	DeleteEvent(ctx context.Context, id int64) error
}

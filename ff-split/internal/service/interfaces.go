package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// CategoryServiceInterface интерфейс для работы с категориями на уровне бизнес-логики
type CategoryServiceInterface interface {
	GetCategories(ctx context.Context, categoryType string) ([]models.EventCategory, error)
	GetCategoryByID(ctx context.Context, id int, categoryType string) (*models.EventCategory, error)
	CreateCategory(ctx context.Context, category *models.EventCategory, categoryType string) (*models.EventCategory, error)
	UpdateCategory(ctx context.Context, id int, category *models.EventCategory, categoryType string) (*models.EventCategory, error)
	DeleteCategory(ctx context.Context, id int, categoryType string) error
	GetCategoryTypes(ctx context.Context) ([]string, error)
}

// EventServiceInterface интерфейс для работы с мероприятиями на уровне бизнес-логики
type EventServiceInterface interface {
	GetEvents(ctx context.Context) ([]models.Event, error)
	GetEventByID(ctx context.Context, id int64) (*models.Event, error)
	CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error)
	UpdateEvent(ctx context.Context, id int64, event *models.Event) (*models.Event, error)
	DeleteEvent(ctx context.Context, id int64) error
}

// ActivityServiceInterface интерфейс для работы с активностями на уровне бизнес-логики
type ActivityServiceInterface interface {
	GetActivitiesByEventID(ctx context.Context, eventID int64) ([]models.Activity, error)
	GetActivityByID(ctx context.Context, id int) (*models.Activity, error)
	CreateActivity(ctx context.Context, activity *models.Activity) (*models.Activity, error)
	UpdateActivity(ctx context.Context, id int, activity *models.Activity) (*models.Activity, error)
	DeleteActivity(ctx context.Context, id int) error
}

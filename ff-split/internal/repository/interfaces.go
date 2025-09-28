package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// CategoryRepository интерфейс для работы с категориями
type CategoryRepository interface {
	GetAll(ctx context.Context, categoryType string) ([]models.EventCategory, error)
	GetByID(ctx context.Context, id int, categoryType string) (*models.EventCategory, error)
	Create(ctx context.Context, category *models.EventCategory, categoryType string) (*models.EventCategory, error)
	Update(ctx context.Context, id int, category *models.EventCategory, categoryType string) (*models.EventCategory, error)
	Delete(ctx context.Context, id int, categoryType string) error
	GetCategoryTypes(ctx context.Context) ([]string, error)
}

// EventRepository интерфейс для работы с мероприятиями
type EventRepository interface {
	GetAll(ctx context.Context) ([]models.Event, error)
	GetByID(ctx context.Context, id int64) (*models.Event, error)
	Create(ctx context.Context, event *models.Event) (*models.Event, error)
	Update(ctx context.Context, id int64, event *models.Event) (*models.Event, error)
	Delete(ctx context.Context, id int64) error
}

// ActivityRepository интерфейс для работы с активностями
type ActivityRepository interface {
	GetByEventID(ctx context.Context, eventID int64) ([]models.Activity, error)
	GetByID(ctx context.Context, id int) (*models.Activity, error)
	Create(ctx context.Context, activity *models.Activity) (*models.Activity, error)
	Update(ctx context.Context, id int, activity *models.Activity) (*models.Activity, error)
	Delete(ctx context.Context, id int) error
}

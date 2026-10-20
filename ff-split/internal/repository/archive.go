package repository

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Archive сериализует смену статуса и предоставляет список доступного архива.
type Archive interface {
	List(context.Context, int64) ([]models.Event, error)
	InEvent(context.Context, int64, func(context.Context, *models.Event) error) error
	IsMember(context.Context, int64, int64) (bool, error)
	SetArchived(context.Context, *models.Event, bool) error
}

package repository

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Invitation отделяет SQL и блокировки ссылок от генерации секрета и правил сервиса.
type Invitation interface {
	InEvent(context.Context, int64, func(context.Context) error) error
	IsMember(context.Context, int64, int64) (bool, error)
	List(context.Context, int64, int64) ([]models.Invitation, error)
	Create(context.Context, *models.Invitation) error
	Get(context.Context, int64, int64) (*models.Invitation, error)
	Lookup(context.Context, string) (*models.Invitation, error)
	EventName(context.Context, int64) (string, error)
	Revoke(context.Context, *models.Invitation) error
	Join(context.Context, int64, int64) error
}

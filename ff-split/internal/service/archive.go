package service

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Archive сохраняет данные события до окончания срока восстановления.
type Archive interface {
	List(context.Context) ([]models.Event, error)
	Archive(context.Context, int64) error
	Restore(context.Context, int64) error
}

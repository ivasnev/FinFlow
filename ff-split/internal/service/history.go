package service

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// History предоставляет только чтение неизменяемой истории.
type History interface {
	List(context.Context, int64, int64) ([]models.HistoryEntry, error)
}

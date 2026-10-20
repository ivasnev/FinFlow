package service

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// SettlementRequest задаёт подтверждённый возврат и ключ безопасного повтора.
type SettlementRequest struct {
	FromUserID  int64  `json:"from_user_id"`
	AmountCents int64  `json:"amount_cents"`
	RequestID   string `json:"request_id"`
}

// Settlement предоставляет учёт возвратов, независимый от истории расходов.
type Settlement interface {
	// List возвращает страницу истории возвратов, включая отменённые записи.
	List(context.Context, int64, int64) ([]models.Settlement, error)
	// Receive подтверждает получение денег текущим пользователем.
	Receive(context.Context, int64, int64, SettlementRequest) (*models.Settlement, error)
	// Cancel отменяет ошибочную запись только от имени её создателя.
	Cancel(context.Context, int64, int64, int64) error
}

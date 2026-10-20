package models

import "time"

// Settlement хранит подтверждённый получателем возврат в целых копейках.
type Settlement struct {
	ID          int64      `json:"id"`
	EventID     int64      `json:"event_id"`
	FromUserID  int64      `json:"from_user_id"`
	ToUserID    int64      `json:"to_user_id"`
	CreatedBy   int64      `json:"created_by"`
	AmountCents int64      `json:"amount_cents"`
	RequestID   string     `json:"request_id"`
	CreatedAt   time.Time  `json:"created_at"`
	CanceledAt  *time.Time `json:"canceled_at,omitempty"`
}

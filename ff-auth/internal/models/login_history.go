package models

import (
	"time"
)

// LoginHistory представляет историю входов пользователя
type LoginHistory struct {
	ID        int       `json:"id"`
	UserID    int64     `json:"user_id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

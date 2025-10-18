package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Session представляет активную сессию пользователя
type Session struct {
	ID           uuid.UUID      `json:"id"`
	UserID       int64          `json:"user_id"`
	RefreshToken string         `json:"-"`
	IPAddress    pq.StringArray `json:"ip_address"`
	ExpiresAt    time.Time      `json:"expires_at"`
	CreatedAt    time.Time      `json:"created_at"`
}

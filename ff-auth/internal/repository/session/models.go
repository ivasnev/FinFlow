package session

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// Session представляет активную сессию пользователя
type Session struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	UserID       int64          `gorm:"type:bigint;not null;column:user_id" json:"user_id"`
	RefreshToken string         `gorm:"type:text;unique;not null;column:refresh_token" json:"-"`
	IPAddress    pq.StringArray `gorm:"type:inet;column:ip_address" json:"ip_address"`
	ExpiresAt    time.Time      `gorm:"type:timestamp;not null;column:expires_at" json:"expires_at"`
	CreatedAt    time.Time      `gorm:"type:timestamp;not null;default:now();column:created_at" json:"created_at"`
}

// TableName устанавливает имя таблицы для модели Session
func (Session) TableName() string {
	return "sessions"
}

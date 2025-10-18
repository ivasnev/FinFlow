package login_history

import (
	"time"
)

// LoginHistory представляет историю входов пользователя
type LoginHistory struct {
	ID        int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64     `gorm:"type:bigint;not null;column:user_id" json:"user_id"`
	IPAddress string    `gorm:"type:inet;not null;column:ip_address" json:"ip_address"`
	UserAgent string    `gorm:"type:text;column:user_agent" json:"user_agent"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now();column:created_at" json:"created_at"`
}

// TableName устанавливает имя таблицы для модели LoginHistory
func (LoginHistory) TableName() string {
	return "login_history"
}

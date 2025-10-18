package device

import (
	"time"
)

// Device представляет устройство, с которого пользователь входил в систему
type Device struct {
	ID        int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64     `gorm:"type:bigint;not null;column:user_id" json:"user_id"`
	DeviceID  string    `gorm:"type:text;unique;not null;column:device_id" json:"device_id"`
	UserAgent string    `gorm:"type:text;not null;column:user_agent" json:"user_agent"`
	LastLogin time.Time `gorm:"type:timestamp;not null;default:now();column:last_login" json:"last_login"`
}

// TableName устанавливает имя таблицы для модели Device
func (Device) TableName() string {
	return "devices"
}

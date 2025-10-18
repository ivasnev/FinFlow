package models

import (
	"time"
)

// Device представляет устройство, с которого пользователь входил в систему
type Device struct {
	ID        int       `json:"id"`
	UserID    int64     `json:"user_id"`
	DeviceID  string    `json:"device_id"`
	UserAgent string    `json:"user_agent"`
	LastLogin time.Time `json:"last_login"`
}

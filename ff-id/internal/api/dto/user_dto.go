package dto

import (
	"time"

	"github.com/google/uuid"
)

// UserDTO представляет данные пользователя, возвращаемые в API
type UserDTO struct {
	ID        int64      `json:"id"`
	Email     string     `json:"email"`
	Phone     *string    `json:"phone,omitempty"`
	Nickname  string     `json:"nickname"`
	Name      *string    `json:"name,omitempty"`
	Birthdate *time.Time `json:"birthdate,omitempty"`
	AvatarID  *uuid.UUID `json:"avatar_id,omitempty"`
	Roles     []string   `json:"roles"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// UpdateUserRequest представляет запрос на обновление данных пользователя
type UpdateUserRequest struct {
	Email     *string    `json:"email,omitempty" binding:"omitempty,email"`
	Phone     *string    `json:"phone,omitempty" binding:"omitempty,e164"`
	Name      *string    `json:"name,omitempty"`
	Birthdate *time.Time `json:"birthdate,omitempty"`
	Password  *string    `json:"password,omitempty" binding:"omitempty,min=8"`
}

// SessionDTO представляет данные о сессии пользователя
type SessionDTO struct {
	ID        uuid.UUID `json:"id"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LoginHistoryDTO представляет данные о входе пользователя
type LoginHistoryDTO struct {
	ID        int       `json:"id"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// DeviceDTO представляет данные об устройстве пользователя
type DeviceDTO struct {
	ID        int       `json:"id"`
	DeviceID  string    `json:"device_id"`
	UserAgent string    `json:"user_agent"`
	LastLogin time.Time `json:"last_login"`
}

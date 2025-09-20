package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// UserRegistration DTO для регистрации пользователя
type UserRegistration struct {
	Email     string        `json:"email" binding:"required,email"`
	Password  string        `json:"password" binding:"required,min=8"`
	Nickname  string        `json:"nickname" binding:"required"`
	Name      string        `json:"name,omitempty" binding:"required"`
	Phone     string        `json:"phone,omitempty"`
	Birthdate string        `json:"birthdate,omitempty" time_format:"2006-01-02"`
	AvatarID  uuid.NullUUID `json:"avatar_id,omitempty"`
}

// UserCredentials DTO для авторизации пользователя
type UserCredentials struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserUpdate DTO для обновления данных пользователя
type UserUpdate struct {
	Nickname  string         `json:"nickname,omitempty"`
	Name      sql.NullString `json:"name,omitempty"`
	Phone     sql.NullString `json:"phone,omitempty"`
	Birthdate sql.NullTime   `json:"birthdate,omitempty"`
	AvatarID  uuid.NullUUID  `json:"avatar_id,omitempty"`
}

// DeviceInfo информация об устройстве пользователя
type DeviceInfo struct {
	UserAgent string `json:"user_agent"`
	IPAddress string `json:"ip_address"`
}

// TokenPair пара токенов - access и refresh
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// TokenClaims данные, хранящиеся в JWT токене
type TokenClaims struct {
	UserID  int64    `json:"user_id"`
	Email   string   `json:"email"`
	Roles   []string `json:"roles"`
	IsAdmin bool     `json:"is_admin"`
}

// SessionInfo информация о сессии пользователя
type SessionInfo struct {
	ID        uuid.UUID `json:"id"`
	Device    string    `json:"device"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

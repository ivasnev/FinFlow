package models

import (
	"database/sql"

	"github.com/google/uuid"
)

// UserRegistration DTO для регистрации пользователя
type UserRegistration struct {
	Email     string        `json:"email" binding:"required,email"`
	Nickname  string        `json:"nickname" binding:"required"`
	Name      string        `json:"name,omitempty" binding:"required"`
	Phone     string        `json:"phone,omitempty"`
	Birthdate string        `json:"birthdate,omitempty" time_format:"2006-01-02"`
	AvatarID  uuid.NullUUID `json:"avatar_id,omitempty"`
}

// UserUpdate DTO для обновления данных пользователя
type UserUpdate struct {
	Nickname  string         `json:"nickname,omitempty"`
	Name      sql.NullString `json:"name,omitempty"`
	Phone     sql.NullString `json:"phone,omitempty"`
	Birthdate sql.NullTime   `json:"birthdate,omitempty"`
	AvatarID  uuid.NullUUID  `json:"avatar_id,omitempty"`
}

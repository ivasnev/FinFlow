package models

import (
	"time"
)

// User представляет пользователя системы
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Nickname     string    `json:"nickname"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Связи
	Roles        []UserRole     `json:"roles,omitempty"`
	Sessions     []Session      `json:"sessions,omitempty"`
	LoginHistory []LoginHistory `json:"login_history,omitempty"`
	Devices      []Device       `json:"devices,omitempty"`
}

package dto

import (
	"github.com/google/uuid"
)

// UserDTO представляет данные пользователя, возвращаемые в API
type UserDTO struct {
	ID        int64      `json:"id"`
	Email     string     `json:"email"`
	Phone     *string    `json:"phone,omitempty"`
	Nickname  string     `json:"nickname"`
	Name      *string    `json:"name,omitempty"`
	Birthdate *int64     `json:"birthdate,omitempty"`
	AvatarID  *uuid.UUID `json:"avatar_id,omitempty"`
	CreatedAt int64      `json:"created_at"`
	UpdatedAt int64      `json:"updated_at"`
}

// ShortUserDTO представляет основные данные пользователя, возвращаемые в API
type ShortUserDTO struct {
	ID       int64   `json:"id"`
	Email    string  `json:"email"`
	Nickname string  `json:"nickname"`
	Name     *string `json:"name,omitempty"`
}

// UpdateUserRequest представляет запрос на обновление данных пользователя
type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty" binding:"omitempty,email"`
	Phone     *string `json:"phone,omitempty" binding:"omitempty,e164"`
	Name      *string `json:"name,omitempty"`
	Birthdate *int64  `json:"birthdate,omitempty" binding:"omitempty"`
	Nickname  *string `json:"nickname,omitempty"`
}

// RegisterUserRequest представляет запрос на регистрацию пользователя
type RegisterUserRequest struct {
	Email     string     `json:"email" binding:"required,email"`
	Nickname  string     `json:"nickname" binding:"required"`
	Name      string     `json:"name,omitempty"`
	Phone     *string    `json:"phone,omitempty" binding:"omitempty,e164"`
	Birthdate *int64     `json:"birthdate,omitempty" binding:"omitempty"`
	AvatarID  *uuid.UUID `json:"avatar_id,omitempty"`
}

// ServiceRegisterUserRequest представляет запрос на регистрацию пользователя от другого сервиса
type ServiceRegisterUserRequest struct {
	UserID   int64  `json:"user_id" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Nickname string `json:"nickname" binding:"required"`
}

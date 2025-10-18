package service

import (
	"context"
	"time"
)

// UserData представляет данные пользователя, возвращаемые в API
type UserData struct {
	Id        int64
	Email     string
	Nickname  string
	Roles     []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserUpdateData представляет запрос на обновление данных пользователя
type UserUpdateData struct {
	Email    *string
	Nickname *string
	Password *string
}

// User определяет методы для работы с пользователями
type User interface {
	// GetUserByID получает пользователя по ID
	GetUserByID(ctx context.Context, id int64) (*UserData, error)

	// GetUserByNickname получает пользователя по никнейму
	GetUserByNickname(ctx context.Context, nickname string) (*UserData, error)

	// UpdateUser обновляет данные пользователя
	UpdateUser(ctx context.Context, userID int64, req UserUpdateData) (*UserData, error)

	// DeleteUser удаляет пользователя
	DeleteUser(ctx context.Context, userID int64) error
}

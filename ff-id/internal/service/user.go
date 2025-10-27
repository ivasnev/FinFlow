package service

import (
	"context"

	"github.com/google/uuid"
)

// UserDTO представляет данные пользователя, возвращаемые в API
type UserDTO struct {
	ID        int64
	Email     string
	Phone     *string
	Nickname  string
	Name      *string
	Birthdate *int64
	AvatarID  *uuid.UUID
	CreatedAt int64
	UpdatedAt int64
}

// ShortUserDTO представляет основные данные пользователя, возвращаемые в API
type ShortUserDTO struct {
	ID       int64
	Email    string
	Nickname string
	Name     *string
}

// UpdateUserRequest представляет запрос на обновление данных пользователя
type UpdateUserRequest struct {
	Email     *string
	Phone     *string
	Name      *string
	Birthdate *int64
	Nickname  *string
}

// RegisterUserRequest представляет запрос на регистрацию пользователя
type RegisterUserRequest struct {
	Email     string
	Nickname  string
	Name      *string
	Phone     *string
	Birthdate *int64
	AvatarID  *uuid.UUID
}

// ServiceRegisterUserRequest представляет запрос на регистрацию пользователя от другого сервиса
type ServiceRegisterUserRequest struct {
	UserID   int64
	Email    string
	Nickname string
}

// UserServiceInterface определяет методы для работы с пользователями
type UserServiceInterface interface {
	// GetUserByID получает пользователя по ID
	GetUserByID(ctx context.Context, id int64) (*UserDTO, error)

	// GetUsersByIds получает пользователей по их ID
	GetUsersByIds(ctx context.Context, ids []int64) ([]*UserDTO, error)

	// GetUserByNickname получает пользователя по никнейму
	GetUserByNickname(ctx context.Context, nickname string) (*UserDTO, error)

	// UpdateUser обновляет данные пользователя
	UpdateUser(ctx context.Context, userID int64, req UpdateUserRequest) (*UserDTO, error)

	// ChangeAvatar изменяет аватар пользователя
	ChangeAvatar(ctx context.Context, userID int64, fileID uuid.UUID) error

	// DeleteUser удаляет пользователя
	DeleteUser(ctx context.Context, userID int64) error

	// RegisterUser регистрирует нового пользователя
	RegisterUser(ctx context.Context, userID int64, user *RegisterUserRequest) (*UserDTO, error)
}

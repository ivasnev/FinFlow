package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/api/dto"
)

// UserServiceInterface определяет методы для работы с пользователями
type UserServiceInterface interface {
	// GetUserByID получает пользователя по ID
	GetUserByID(ctx context.Context, id int64) (*dto.UserDTO, error)

	// GetUserByNickname получает пользователя по никнейму
	GetUserByNickname(ctx context.Context, nickname string) (*dto.UserDTO, error)

	// UpdateUser обновляет данные пользователя
	UpdateUser(ctx context.Context, userID int64, req dto.UpdateUserRequest) (*dto.UserDTO, error)

	// ChangeAvatar изменяет аватар пользователя
	ChangeAvatar(ctx context.Context, userID int64, fileID uuid.UUID) error

	// DeleteUser удаляет пользователя
	DeleteUser(ctx context.Context, userID int64) error

	// RegisterUser регистрирует нового пользователя
	RegisterUser(ctx context.Context, userID int64, user *dto.RegisterUserRequest) (*dto.UserDTO, error)
}

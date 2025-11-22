package adapters

import (
	"context"
)

// IDAdapter определяет интерфейс для работы с ff-id сервисом
type IDAdapter interface {
	GetUserByID(ctx context.Context, userID int64) (*UserDTO, error)
	GetUsersByIDs(ctx context.Context, userIDs []int64) ([]UserDTO, error)
}

// UserDTO - данные пользователя из ff-id сервиса
type UserDTO struct {
	ID        int64
	Email     string
	Nickname  string
	Name      *string
	Phone     *string
	AvatarID  *string
	Birthdate *int64
	CreatedAt int64
	UpdatedAt int64
}

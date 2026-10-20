package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
)

// Identity обращается к единственному источнику email и nickname в ff-id.
type Identity interface {
	// RegisterUser создаёт профиль с ID учётных данных.
	RegisterUser(ctx context.Context, req *ffid.RegisterUserRequest) (*ffid.UserDTO, error)
	// FindIdentity находит профиль с явно заданной семантикой login, email либо nickname.
	FindIdentity(ctx context.Context, login string, kind string) (*ffid.UserDTO, error)
	// GetIdentity получает актуальный профиль по стабильному ID.
	GetIdentity(ctx context.Context, id int64) (*ffid.UserDTO, error)
	// UpdateIdentity изменяет идентификаторы в их единственном хранилище.
	UpdateIdentity(ctx context.Context, id int64, email, nickname *string) (*ffid.UserDTO, error)
}

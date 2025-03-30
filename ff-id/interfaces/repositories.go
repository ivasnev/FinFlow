package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// UserRepository интерфейс для работы с пользователями
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByNickname(ctx context.Context, nickname string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int64) error
	AddRole(ctx context.Context, userID int64, roleID int) error
	RemoveRole(ctx context.Context, userID int64, roleID int) error
	GetRoles(ctx context.Context, userID int64) ([]models.RoleEntity, error)
}

// RoleRepository интерфейс для работы с ролями
type RoleRepository interface {
	GetByName(ctx context.Context, name string) (*models.RoleEntity, error)
	GetAll(ctx context.Context) ([]models.RoleEntity, error)
}

// SessionRepository интерфейс для работы с сессиями
type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error)
	GetAllByUserID(ctx context.Context, userID int64) ([]models.Session, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID int64) error
	DeleteExpired(ctx context.Context) error
}

// LoginHistoryRepository интерфейс для работы с историей входов
type LoginHistoryRepository interface {
	Create(ctx context.Context, history *models.LoginHistory) error
	GetAllByUserID(ctx context.Context, userID int64, limit, offset int) ([]models.LoginHistory, error)
}

// DeviceRepository интерфейс для работы с устройствами
type DeviceRepository interface {
	Create(ctx context.Context, device *models.Device) error
	GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error)
	GetAllByUserID(ctx context.Context, userID int64) ([]models.Device, error)
	Update(ctx context.Context, device *models.Device) error
	UpdateLastLogin(ctx context.Context, deviceID string, lastLogin time.Time) error
	Delete(ctx context.Context, id int) error
}

// AvatarRepository интерфейс для работы с аватарками
type AvatarRepository interface {
	Create(ctx context.Context, avatar *models.UserAvatar) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.UserAvatar, error)
	GetAllByUserID(ctx context.Context, userID int64) ([]models.UserAvatar, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteAllByUserID(ctx context.Context, userID int64) error
}

package interfaces

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// UserRepository определяет методы для работы с пользователями
type UserRepository interface {
	// Create создает нового пользователя
	Create(ctx context.Context, user *models.User) error

	// GetByID находит пользователя по ID
	GetByID(ctx context.Context, id int64) (*models.User, error)

	// GetByEmail находит пользователя по email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// GetByNickname находит пользователя по никнейму
	GetByNickname(ctx context.Context, nickname string) (*models.User, error)

	// Update обновляет данные пользователя
	Update(ctx context.Context, user *models.User) error

	// Delete удаляет пользователя
	Delete(ctx context.Context, id int64) error

	// AddRole добавляет пользователю роль
	AddRole(ctx context.Context, userID int64, roleID int) error

	// RemoveRole удаляет роль у пользователя
	RemoveRole(ctx context.Context, userID int64, roleID int) error

	// GetRoles получает все роли пользователя
	GetRoles(ctx context.Context, userID int64) ([]models.RoleEntity, error)
}

// RoleRepository определяет методы для работы с ролями
type RoleRepository interface {
	// GetByName находит роль по имени
	GetByName(ctx context.Context, name string) (*models.RoleEntity, error)

	// GetAll получает все роли
	GetAll(ctx context.Context) ([]models.RoleEntity, error)
}

// SessionRepository определяет методы для работы с сессиями
type SessionRepository interface {
	// Create создает новую сессию
	Create(ctx context.Context, session *models.Session) error

	// GetByID находит сессию по ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error)

	// GetByRefreshToken находит сессию по refresh-токену
	GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error)

	// GetAllByUserID получает все сессии пользователя
	GetAllByUserID(ctx context.Context, userID int64) ([]models.Session, error)

	// Delete удаляет сессию
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteAllByUserID удаляет все сессии пользователя
	DeleteAllByUserID(ctx context.Context, userID int64) error

	// DeleteExpired удаляет все истекшие сессии
	DeleteExpired(ctx context.Context) error
}

// LoginHistoryRepository определяет методы для работы с историей входов
type LoginHistoryRepository interface {
	// Create создает новую запись в истории входов
	Create(ctx context.Context, history *models.LoginHistory) error

	// GetAllByUserID получает всю историю входов пользователя
	GetAllByUserID(ctx context.Context, userID int64, limit, offset int) ([]models.LoginHistory, error)
}

// DeviceRepository определяет методы для работы с устройствами
type DeviceRepository interface {
	// Create создает новое устройство
	Create(ctx context.Context, device *models.Device) error

	// GetByDeviceID находит устройство по deviceID
	GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error)

	// GetAllByUserID получает все устройства пользователя
	GetAllByUserID(ctx context.Context, userID int64) ([]models.Device, error)

	// Update обновляет данные устройства
	Update(ctx context.Context, device *models.Device) error

	// UpdateLastLogin обновляет время последнего входа
	UpdateLastLogin(ctx context.Context, deviceID string, lastLogin time.Time) error

	// Delete удаляет устройство
	Delete(ctx context.Context, id int) error
}

// AvatarRepository определяет методы для работы с аватарками
type AvatarRepository interface {
	// Create создает новую аватарку
	Create(ctx context.Context, avatar *models.UserAvatar) error

	// GetByID находит аватарку по ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.UserAvatar, error)

	// GetAllByUserID получает все аватарки пользователя
	GetAllByUserID(ctx context.Context, userID int64) ([]models.UserAvatar, error)

	// Delete удаляет аватарку
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteAllByUserID удаляет все аватарки пользователя
	DeleteAllByUserID(ctx context.Context, userID int64) error
}

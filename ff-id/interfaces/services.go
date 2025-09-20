package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// AuthService интерфейс для сервиса аутентификации
type AuthService interface {
	Register(ctx context.Context, user *models.UserRegistration) (*models.User, error)
	Login(ctx context.Context, credentials *models.UserCredentials, deviceInfo *models.DeviceInfo) (*models.TokenPair, error)
	RefreshTokens(ctx context.Context, refreshToken string, deviceInfo *models.DeviceInfo) (*models.TokenPair, error)
	Logout(ctx context.Context, userID int64, refreshToken string) error
	ValidateToken(token string) (*models.TokenClaims, error)
	HasRole(ctx context.Context, userID int64, roleName string) (bool, error)
}

// UserService интерфейс для сервиса пользователей
type UserService interface {
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetByNickname(ctx context.Context, nickname string) (*models.User, error)
	Update(ctx context.Context, id int64, userData *models.UserUpdate) (*models.User, error)
	UpdateAvatar(ctx context.Context, userID int64, avatarID uuid.UUID) error
	DeleteAvatar(ctx context.Context, userID int64, avatarID uuid.UUID) error
}

// SessionService интерфейс для сервиса сессий
type SessionService interface {
	GetAllByUserID(ctx context.Context, userID int64) ([]models.SessionInfo, error)
	TerminateSession(ctx context.Context, userID int64, sessionID uuid.UUID) error
	TerminateAllSessions(ctx context.Context, userID int64) error
}

// LoginHistoryService интерфейс для сервиса истории входов
type LoginHistoryService interface {
	GetByUserID(ctx context.Context, userID int64, page, pageSize int) ([]models.LoginHistory, error)
}

// DeviceService интерфейс для сервиса устройств
type DeviceService interface {
	GetOrCreate(ctx context.Context, userID int64, deviceInfo *models.DeviceInfo) (*models.Device, error)
	UpdateLastLogin(ctx context.Context, deviceID string) error
	GetAllByUserID(ctx context.Context, userID int64) ([]models.Device, error)
}

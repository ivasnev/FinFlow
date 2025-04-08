package service

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// AuthServiceInterface определяет методы для аутентификации и авторизации
type AuthServiceInterface interface {
	// Register регистрирует нового пользователя
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)

	// Login выполняет вход пользователя в систему
	Login(ctx context.Context, req dto.LoginRequest, r *http.Request) (*dto.AuthResponse, error)

	// RefreshToken обновляет access-токен
	RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)

	// Logout выполняет выход пользователя из системы
	Logout(ctx context.Context, refreshToken string) error

	// GenerateTokenPair генерирует пару токенов (access и refresh)
	GenerateTokenPair(ctx context.Context, userID int64, roles []string) (accessToken, refreshToken string, expiresAt int64, err error)

	// ValidateToken проверяет валидность токена
	ValidateToken(token string) (int64, []string, error)

	// RecordLogin записывает историю входа
	RecordLogin(ctx context.Context, userID int64, r *http.Request) error
}

// UserServiceInterface определяет методы для работы с пользователями
type UserServiceInterface interface {
	// GetUserByID получает пользователя по ID
	GetUserByID(ctx context.Context, id int64) (*dto.UserDTO, error)

	// GetUserByNickname получает пользователя по никнейму
	GetUserByNickname(ctx context.Context, nickname string) (*dto.UserDTO, error)

	// UpdateUser обновляет данные пользователя
	UpdateUser(ctx context.Context, userID int64, req dto.UpdateUserRequest) (*dto.UserDTO, error)

	// DeleteUser удаляет пользователя
	DeleteUser(ctx context.Context, userID int64) error
}

// SessionServiceInterface определяет методы для работы с сессиями
type SessionServiceInterface interface {
	// GetUserSessions получает все сессии пользователя
	GetUserSessions(ctx context.Context, userID int64) ([]dto.SessionDTO, error)

	// TerminateSession завершает сессию
	TerminateSession(ctx context.Context, sessionID uuid.UUID, userID int64) error

	// TerminateAllSessions завершает все сессии пользователя
	TerminateAllSessions(ctx context.Context, userID int64) error
}

// LoginHistoryServiceInterface определяет методы для работы с историей входов
type LoginHistoryServiceInterface interface {
	// GetUserLoginHistory получает историю входов пользователя
	GetUserLoginHistory(ctx context.Context, userID int64, limit, offset int) ([]dto.LoginHistoryDTO, error)
}

// DeviceServiceInterface определяет методы для работы с устройствами
type DeviceServiceInterface interface {
	// GetUserDevices получает все устройства пользователя
	GetUserDevices(ctx context.Context, userID int64) ([]dto.DeviceDTO, error)

	// RemoveDevice удаляет устройство
	RemoveDevice(ctx context.Context, deviceID int, userID int64) error

	// GetOrCreateDevice получает или создает устройство по deviceID
	GetOrCreateDevice(ctx context.Context, deviceID string, userAgent string, userID int64) (*models.Device, error)
}

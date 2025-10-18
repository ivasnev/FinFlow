package service

import (
	"context"
	"crypto/ed25519"
	"time"
)

// RegisterParams представляет запрос на регистрацию нового пользователя
type RegisterParams struct {
	Email    string
	Phone    *string
	Password string
	Nickname string
	Name     *string
}

// LoginParams представляет запрос на вход в систему
type LoginParams struct {
	Login     string // Может быть email или nickname
	Password  string
	UserAgent string
	IpAddress string
}

// RefreshTokenParams представляет запрос на обновление access-токена
type RefreshTokenParams struct {
	RefreshToken string
}

// LogoutParams представляет запрос на выход из системы
type LogoutParams struct {
	RefreshToken string
}

// AuthResponse представляет ответ после успешной аутентификации
type AccessDataParams struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	User         ShortUserParams
}

// ShortUserParams представляет основные данные пользователя, возвращаемые в API
type ShortUserParams struct {
	Id       int64
	Email    string
	Nickname string
	Roles    []string
}

// Auth определяет методы для аутентификации и авторизации
type Auth interface {
	// Register регистрирует нового пользователя
	Register(ctx context.Context, req RegisterParams) (*AccessDataParams, error)

	// Login выполняет вход пользователя в систему
	Login(ctx context.Context, req LoginParams) (*AccessDataParams, error)

	// RefreshToken обновляет access-токен
	RefreshToken(ctx context.Context, refreshToken string) (*AccessDataParams, error)

	// Logout выполняет выход пользователя из системы
	Logout(ctx context.Context, refreshToken string) error

	// GenerateTokenPair генерирует пару токенов (access и refresh)
	GenerateTokenPair(ctx context.Context, userID int64, roles []string) (accessToken, refreshToken string, expiresAt int64, err error)

	// ValidateToken проверяет валидность токена
	ValidateToken(token string) (int64, []string, error)

	// RecordLogin записывает историю входа
	RecordLogin(ctx context.Context, userID int64, ipAddress string, userAgent string) error

	// GetPublicKey возвращает публичный ключ для проверки токенов
	GetPublicKey() ed25519.PublicKey
}

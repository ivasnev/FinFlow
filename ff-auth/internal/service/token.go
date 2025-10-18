package service

import (
	"crypto/ed25519"
	"time"
)

// TokenPayload представляет содержимое токена
type TokenPayload struct {
	UserID int64    `json:"user_id"`
	Roles  []string `json:"roles"`
	Exp    int64    `json:"exp"`
}

// Token представляет структуру токена
type Token struct {
	Payload []byte `json:"payload"`
	Sig     []byte `json:"sig"`
}

// TokenManager определяет методы для работы с токенами
type TokenManager interface {
	// GetPublicKey возвращает публичный ключ для проверки токенов
	LoadOrGenerateKeys() error
	// RegenerateKeys создает новую пару ключей для подписи токенов и сохраняет в БД
	RegenerateKeys() error
	// GetPublicKey возвращает текущий публичный ключ
	GetPublicKey() ed25519.PublicKey
	// GenerateToken создает новый токен
	GenerateToken(payload *TokenPayload) (string, error)
	// ValidateToken проверяет валидность токена
	ValidateToken(tokenStr string) (*TokenPayload, error)
	// GenerateTokenPair генерирует пару токенов: access и refresh
	GenerateTokenPair(userID int64, roles []string, accessTTL, refreshTTL time.Duration) (accessToken, refreshToken string, accessExpiresAt int64, err error)
}

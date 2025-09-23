package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	"sync"
	"time"
)

// TokenValidator определяет интерфейс для валидации токенов
type TokenValidator interface {
	ValidateToken(tokenStr string) (*auth.TokenPayload, error)
	GetPublicKey() ed25519.PublicKey
}

// TokenGenerator определяет интерфейс для генерации токенов
type TokenGenerator interface {
	GenerateToken(payload *auth.TokenPayload) (string, error)
}

// ED25519TokenManager реализует TokenValidator и TokenGenerator с использованием Ed25519
type ED25519TokenManager struct {
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
	mutex      sync.RWMutex
}

// NewED25519TokenManager создает новый менеджер токенов с использованием Ed25519
func NewED25519TokenManager() (*ED25519TokenManager, error) {
	// Генерируем новую пару ключей
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &ED25519TokenManager{
		publicKey:  publicKey,
		privateKey: privateKey,
	}, nil
}

// RegenerateKeys создает новую пару ключей для подписи токенов
func (m *ED25519TokenManager) RegenerateKeys() error {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.publicKey = publicKey
	m.privateKey = privateKey

	return nil
}

// GetPublicKey возвращает текущий публичный ключ
func (m *ED25519TokenManager) GetPublicKey() ed25519.PublicKey {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.publicKey
}

// GenerateToken создает новый токен
func (m *ED25519TokenManager) GenerateToken(payload *auth.TokenPayload) (string, error) {
	// Сериализуем payload в JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Подписываем payload
	m.mutex.RLock()
	signature := ed25519.Sign(m.privateKey, payloadBytes)
	m.mutex.RUnlock()

	// Формируем структуру токена
	token := auth.Token{
		Payload: payloadBytes,
		Sig:     signature,
	}

	// Сериализуем токен в JSON
	tokenBytes, err := json.Marshal(token)
	if err != nil {
		return "", err
	}

	// Кодируем в base64
	tokenStr := base64.StdEncoding.EncodeToString(tokenBytes)

	return tokenStr, nil
}

// ValidateToken проверяет валидность токена
func (m *ED25519TokenManager) ValidateToken(tokenStr string) (*auth.TokenPayload, error) {
	// Декодируем из base64
	tokenBytes, err := base64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		return nil, errors.New("invalid token format")
	}

	// Десериализуем токен
	var token auth.Token
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return nil, errors.New("invalid token data")
	}

	// Проверяем подпись
	m.mutex.RLock()
	valid := ed25519.Verify(m.publicKey, token.Payload, token.Sig)
	m.mutex.RUnlock()

	if !valid {
		return nil, errors.New("invalid token signature")
	}

	// Десериализуем payload
	var payload auth.TokenPayload
	if err := json.Unmarshal(token.Payload, &payload); err != nil {
		return nil, errors.New("invalid payload data")
	}

	// Проверяем срок действия токена
	if payload.Exp < time.Now().Unix() {
		return nil, errors.New("token expired")
	}

	return &payload, nil
}

// GenerateTokenPair генерирует пару токенов: access и refresh
func (m *ED25519TokenManager) GenerateTokenPair(userID int64, roles []string, accessTTL, refreshTTL time.Duration) (accessToken, refreshToken string, accessExpiresAt int64, err error) {
	// Создаем payload для access токена
	now := time.Now()
	accessExpiresAt = now.Add(accessTTL).Unix()

	accessPayload := &auth.TokenPayload{
		UserID: userID,
		Roles:  roles,
		Exp:    accessExpiresAt,
	}

	// Генерируем access токен
	accessToken, err = m.GenerateToken(accessPayload)
	if err != nil {
		return "", "", 0, err
	}

	// Создаем payload для refresh токена
	refreshExpiresAt := now.Add(refreshTTL).Unix()
	refreshPayload := &auth.TokenPayload{
		UserID: userID,
		Roles:  roles,
		Exp:    refreshExpiresAt,
	}

	// Генерируем refresh токен
	refreshToken, err = m.GenerateToken(refreshPayload)
	if err != nil {
		return "", "", 0, err
	}

	return accessToken, refreshToken, accessExpiresAt, nil
}

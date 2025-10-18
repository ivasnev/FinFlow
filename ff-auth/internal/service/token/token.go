package token

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	pg_repos "github.com/ivasnev/FinFlow/ff-auth/internal/repository/postgres"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
)

// ED25519TokenManager реализует TokenManager с использованием Ed25519
type ED25519TokenManager struct {
	publicKey    ed25519.PublicKey
	privateKey   ed25519.PrivateKey
	mutex        sync.RWMutex
	keyPairRepo  pg_repos.KeyPairRepositoryInterface
	loadedFromDB bool
}

// NewED25519TokenManager создает новый менеджер токенов с использованием Ed25519
func NewED25519TokenManager(keyPairRepo pg_repos.KeyPairRepositoryInterface) (*ED25519TokenManager, error) {
	manager := &ED25519TokenManager{
		keyPairRepo:  keyPairRepo,
		loadedFromDB: false,
	}

	// Загрузка ключей из БД или генерация новых
	if err := manager.LoadOrGenerateKeys(); err != nil {
		return nil, err
	}

	return manager, nil
}

// LoadOrGenerateKeys загружает ключи из БД или генерирует новые, если в БД их нет
func (m *ED25519TokenManager) LoadOrGenerateKeys() error {
	// Попытка загрузить активную пару ключей из БД
	keyPair, err := m.keyPairRepo.GetActive(context.Background())
	if err != nil {
		return fmt.Errorf("ошибка при загрузке ключей: %w", err)
	}

	if keyPair != nil {
		// Ключи найдены в БД, используем их
		publicKeyBytes, err := base64.StdEncoding.DecodeString(keyPair.PublicKey)
		if err != nil {
			return fmt.Errorf("ошибка декодирования публичного ключа: %w", err)
		}

		privateKeyBytes, err := base64.StdEncoding.DecodeString(keyPair.PrivateKey)
		if err != nil {
			return fmt.Errorf("ошибка декодирования приватного ключа: %w", err)
		}

		m.mutex.Lock()
		m.publicKey = ed25519.PublicKey(publicKeyBytes)
		m.privateKey = ed25519.PrivateKey(privateKeyBytes)
		m.loadedFromDB = true
		m.mutex.Unlock()

		log.Println("Ключи успешно загружены из базы данных")
	} else {
		// Ключей в БД нет, генерируем новые
		if err := m.RegenerateKeys(); err != nil {
			return fmt.Errorf("ошибка при генерации ключей: %w", err)
		}
	}

	return nil
}

// RegenerateKeys создает новую пару ключей для подписи токенов и сохраняет в БД
func (m *ED25519TokenManager) RegenerateKeys() error {
	// Генерируем новую пару ключей
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	// Кодируем ключи в base64 для сохранения в БД
	publicKeyStr := base64.StdEncoding.EncodeToString(publicKey)
	privateKeyStr := base64.StdEncoding.EncodeToString(privateKey)

	m.mutex.Lock()
	m.publicKey = publicKey
	m.privateKey = privateKey
	m.mutex.Unlock()

	// Сохраняем ключи в БД
	keyPair := &models.KeyPair{
		PublicKey:  publicKeyStr,
		PrivateKey: privateKeyStr,
		IsActive:   true,
	}

	ctx := context.Background()
	// Если уже есть активный ключ, деактивируем его
	if m.loadedFromDB {
		existing, err := m.keyPairRepo.GetActive(ctx)
		if err != nil {
			return fmt.Errorf("ошибка при получении текущего активного ключа: %w", err)
		}
		if existing != nil {
			existing.IsActive = false
			if err := m.keyPairRepo.Update(ctx, existing); err != nil {
				return fmt.Errorf("ошибка при деактивации текущего ключа: %w", err)
			}
		}
	}

	// Сохраняем новый ключ
	if err := m.keyPairRepo.Create(ctx, keyPair); err != nil {
		return fmt.Errorf("ошибка при сохранении ключей в БД: %w", err)
	}

	m.loadedFromDB = true
	log.Println("Сгенерированы и сохранены новые ключи в базе данных")

	return nil
}

// GetPublicKey возвращает текущий публичный ключ
func (m *ED25519TokenManager) GetPublicKey() ed25519.PublicKey {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.publicKey
}

// GenerateToken создает новый токен
func (m *ED25519TokenManager) GenerateToken(payload *service.TokenPayload) (string, error) {
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
	token := service.Token{
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
func (m *ED25519TokenManager) ValidateToken(tokenStr string) (*service.TokenPayload, error) {
	// Декодируем из base64
	tokenBytes, err := base64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		return nil, errors.New("invalid token format")
	}

	// Десериализуем токен
	var token service.Token
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
	var payload service.TokenPayload
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

	accessPayload := &service.TokenPayload{
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
	refreshPayload := &service.TokenPayload{
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

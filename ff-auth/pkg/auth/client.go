package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"
)

// Client представляет клиент для проверки токенов
type Client struct {
	publicKey      ed25519.PublicKey
	publicKeyURL   string
	mutex          sync.RWMutex
	updateInterval time.Duration
	lastUpdate     time.Time
	httpClient     *http.Client
}

// NewClient создает новый клиент для проверки токенов
func NewClient(publicKeyURL string, updateInterval time.Duration) *Client {
	return &Client{
		publicKeyURL:   publicKeyURL,
		updateInterval: updateInterval,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

// GetPublicKey возвращает текущий публичный ключ
func (c *Client) GetPublicKey() (ed25519.PublicKey, error) {
	c.mutex.RLock()
	if c.publicKey != nil && time.Since(c.lastUpdate) < c.updateInterval {
		key := c.publicKey
		c.mutex.RUnlock()
		return key, nil
	}
	c.mutex.RUnlock()

	// Обновляем ключ если прошло больше updateInterval с момента последнего обновления
	return c.fetchPublicKey()
}

// fetchPublicKey получает актуальный публичный ключ с сервера
func (c *Client) fetchPublicKey() (ed25519.PublicKey, error) {
	resp, err := c.httpClient.Get(c.publicKeyURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get public key")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Предполагаем, что ключ передается как base64-строка
	key, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return nil, err
	}

	c.mutex.Lock()
	c.publicKey = key
	c.lastUpdate = time.Now()
	c.mutex.Unlock()

	return key, nil
}

// ValidateToken проверяет валидность токена
func (c *Client) ValidateToken(tokenStr string) (*TokenPayload, error) {
	// Получаем актуальный публичный ключ
	publicKey, err := c.GetPublicKey()
	if err != nil {
		return nil, err
	}

	// Декодируем из base64
	tokenBytes, err := base64.StdEncoding.DecodeString(tokenStr)
	if err != nil {
		return nil, errors.New("invalid token format")
	}

	// Десериализуем токен
	var token Token
	if err := json.Unmarshal(tokenBytes, &token); err != nil {
		return nil, errors.New("invalid token data")
	}

	// Проверяем подпись с использованием публичного ключа
	valid := ed25519.Verify(publicKey, token.Payload, token.Sig)
	if !valid {
		return nil, errors.New("invalid token signature")
	}

	// Десериализуем payload
	var payload TokenPayload
	if err := json.Unmarshal(token.Payload, &payload); err != nil {
		return nil, errors.New("invalid payload data")
	}

	// Проверяем срок действия токена
	if payload.Exp < time.Now().Unix() {
		return nil, errors.New("token expired")
	}

	return &payload, nil
}

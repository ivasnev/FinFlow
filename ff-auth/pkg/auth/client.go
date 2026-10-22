package auth

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

var (
	publicKeyUrl = "/api/v1/auth/public-key"
)

// Client представляет клиент для проверки токенов
type Client struct {
	publicKey      ed25519.PublicKey
	hostURL        string
	mutex          sync.RWMutex
	updateInterval time.Duration
	lastUpdate     time.Time
	httpClient     *http.Client
}

// NewClient создает новый клиент для проверки токенов
func NewClient(hostURL string, updateInterval time.Duration) *Client {
	return &Client{
		hostURL:        hostURL,
		updateInterval: updateInterval,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
	}
}

// GetPublicKey возвращает текущий публичный ключ
func (c *Client) GetPublicKey() (ed25519.PublicKey, error) {
	return c.getPublicKeyContext(context.Background())
}
func (c *Client) getPublicKeyContext(ctx context.Context) (ed25519.PublicKey, error) {
	c.mutex.RLock()
	if c.publicKey != nil && time.Since(c.lastUpdate) < c.updateInterval {
		key := c.publicKey
		c.mutex.RUnlock()
		return key, nil
	}
	c.mutex.RUnlock()

	// Обновляем ключ если прошло больше updateInterval с момента последнего обновления
	return c.fetchPublicKeyContext(ctx)
}

// fetchPublicKey получает актуальный публичный ключ с сервера
func (c *Client) fetchPublicKeyContext(ctx context.Context) (ed25519.PublicKey, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.hostURL+publicKeyUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Небольшое тело ошибки дочитываем для повторного использования соединения.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
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

// ValidateToken проверяет подпись, срок и назначение access-токена.
func (c *Client) ValidateToken(tokenStr string) (*TokenPayload, error) {
	return c.ValidateTokenContext(context.Background(), tokenStr)
}
func (c *Client) ValidateTokenContext(ctx context.Context, tokenStr string) (*TokenPayload, error) {
	if len(tokenStr) > 8192 {
		return nil, errors.New("invalid token size")
	}
	// Получаем актуальный публичный ключ
	publicKey, err := c.getPublicKeyContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: public key unavailable", ErrUnavailable)
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
	valid := len(publicKey) == ed25519.PublicKeySize && ed25519.Verify(publicKey, token.Payload, token.Sig)
	if !valid {
		return nil, errors.New("invalid token signature")
	}

	// Десериализуем payload
	var payload TokenPayload
	if err := json.Unmarshal(token.Payload, &payload); err != nil {
		return nil, errors.New("invalid payload data")
	}
	if payload.Purpose != "access" {
		return nil, errors.New("access token required")
	}

	// Проверяем срок действия токена
	if payload.Exp <= time.Now().Unix() {
		return nil, errors.New("token expired")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.hostURL+"/api/v1/auth/introspect", bytes.NewReader(nil))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: introspection unavailable", ErrUnavailable)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, errors.New("session revoked or email unverified")
	}
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("%w: introspection status %d", ErrUnavailable, resp.StatusCode)
	}
	var result struct {
		Active bool  `json:"active"`
		UserID int64 `json:"user_id"`
	}
	if err = json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&result); err != nil || !result.Active || result.UserID != payload.UserID {
		return nil, fmt.Errorf("%w: invalid introspection response", ErrUnavailable)
	}
	return &payload, nil
}

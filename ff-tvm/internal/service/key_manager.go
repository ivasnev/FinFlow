package service

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
)

// keyManagerImpl реализует интерфейс KeyManager
type keyManagerImpl struct{}

// NewKeyManager создает новый менеджер ключей
func NewKeyManager() KeyManager {
	return &keyManagerImpl{}
}

// GenerateKeyPair генерирует новую пару ключей
func (m *keyManagerImpl) GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(nil)
}

// Sign подписывает данные с помощью приватного ключа
func (m *keyManagerImpl) Sign(data []byte, privateKey ed25519.PrivateKey) ([]byte, error) {
	return ed25519.Sign(privateKey, data), nil
}

// Verify проверяет подпись с помощью публичного ключа
func (m *keyManagerImpl) Verify(data, signature []byte, publicKey ed25519.PublicKey) bool {
	return ed25519.Verify(publicKey, data, signature)
}

// EncodeKey кодирует ключ в base64
func EncodeKey(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

// DecodeKey декодирует ключ из base64
func DecodeKey(keyStr string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(keyStr)
}

// HashKey хеширует ключ с помощью SHA-256
func HashKey(key []byte) string {
	hash := sha256.Sum256(key)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// VerifyKeyHash проверяет, соответствует ли хеш ключу
func VerifyKeyHash(key []byte, hash string) bool {
	return HashKey(key) == hash
}

package service

import (
	"crypto/ed25519"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKeyPair(t *testing.T) {
	manager := NewKeyManager()

	// Генерируем пару ключей
	publicKey, privateKey, err := manager.GenerateKeyPair()
	assert.NoError(t, err)
	assert.NotNil(t, publicKey)
	assert.NotNil(t, privateKey)
	assert.Equal(t, ed25519.PublicKeySize, len(publicKey))
	assert.Equal(t, ed25519.PrivateKeySize, len(privateKey))
}

func TestSignAndVerify(t *testing.T) {
	manager := NewKeyManager()

	// Генерируем пару ключей
	publicKey, privateKey, err := manager.GenerateKeyPair()
	assert.NoError(t, err)

	// Тестовые данные
	data := []byte("test data")

	// Подписываем данные
	signature, err := manager.Sign(data, privateKey)
	assert.NoError(t, err)
	assert.NotNil(t, signature)
	assert.Equal(t, ed25519.SignatureSize, len(signature))

	// Проверяем подпись
	valid := manager.Verify(data, signature, publicKey)
	assert.True(t, valid)

	// Проверяем подпись с измененными данными
	invalidData := []byte("invalid data")
	valid = manager.Verify(invalidData, signature, publicKey)
	assert.False(t, valid)

	// Проверяем подпись с неправильным ключом
	invalidPublicKey, _, _ := manager.GenerateKeyPair()
	valid = manager.Verify(data, signature, invalidPublicKey)
	assert.False(t, valid)
}

func TestEncodeDecodeKeys(t *testing.T) {
	manager := NewKeyManager()

	// Генерируем пару ключей
	publicKey, privateKey, err := manager.GenerateKeyPair()
	assert.NoError(t, err)

	// Кодируем ключи в base64
	publicKeyStr := EncodeKey(publicKey)
	privateKeyStr := EncodeKey(privateKey)

	// Проверяем, что строки не пустые
	assert.NotEmpty(t, publicKeyStr)
	assert.NotEmpty(t, privateKeyStr)

	// Декодируем ключи обратно
	decodedPublicKey, err := DecodeKey(publicKeyStr)
	assert.NoError(t, err)
	assert.Equal(t, publicKey, decodedPublicKey)

	decodedPrivateKey, err := DecodeKey(privateKeyStr)
	assert.NoError(t, err)
	assert.Equal(t, privateKey, decodedPrivateKey)

	// Проверяем, что декодированные ключи работают
	data := []byte("test data")
	signature, err := manager.Sign(data, ed25519.PrivateKey(decodedPrivateKey))
	assert.NoError(t, err)

	valid := manager.Verify(data, signature, ed25519.PublicKey(decodedPublicKey))
	assert.True(t, valid)
}

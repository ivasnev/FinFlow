package service

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type MockKeyManager struct {
	KeyManager
}

func (m *MockKeyManager) GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(nil)
}

func (m *MockKeyManager) Sign(data []byte, privateKey ed25519.PrivateKey) ([]byte, error) {
	return ed25519.Sign(privateKey, data), nil
}

func (m *MockKeyManager) Verify(data, signature []byte, publicKey ed25519.PublicKey) bool {
	return ed25519.Verify(publicKey, data, signature)
}

type MockAccessManager struct {
	AccessManager
	hasAccess bool
}

func (m *MockAccessManager) CheckAccess(from, to int64) bool {
	return m.hasAccess
}

func (m *MockAccessManager) GrantAccess(from, to int64) error {
	return nil
}

func (m *MockAccessManager) RevokeAccess(from, to int64) error {
	return nil
}

type MockServiceRepository struct {
	ServiceRepository
	service *Service
}

func (m *MockServiceRepository) Create(ctx context.Context, service *Service) error {
	return nil
}

func (m *MockServiceRepository) GetByID(ctx context.Context, id int64) (*Service, error) {
	return m.service, nil
}

func (m *MockServiceRepository) GetPublicKey(ctx context.Context, id int64) (string, error) {
	return m.service.PublicKey, nil
}

func TestGenerateTicket(t *testing.T) {
	// Создаем моки
	keyManager := &MockKeyManager{}
	accessManager := &MockAccessManager{hasAccess: true}

	// Генерируем ключи для тестового сервиса
	publicKey, _, err := keyManager.GenerateKeyPair()
	assert.NoError(t, err)

	repo := &MockServiceRepository{
		service: &Service{
			ID:        1,
			Name:      "test-service",
			PublicKey: EncodeKey(publicKey),
		},
	}

	service := NewTicketService(repo, keyManager, accessManager)

	// Тест: Успешная генерация тикета
	ticket, err := service.GenerateTicket(context.Background(), 1, 2)
	assert.NoError(t, err)
	assert.NotNil(t, ticket)
	assert.Equal(t, int64(1), ticket.From)
	assert.Equal(t, int64(2), ticket.To)
	assert.NotEmpty(t, ticket.Signature)
	assert.NotEmpty(t, ticket.Metadata)

	// Тест: Отказ в доступе
	accessManager.hasAccess = false
	ticket, err = service.GenerateTicket(context.Background(), 1, 2)
	assert.Error(t, err)
	assert.Equal(t, ErrAccessDenied, err)
}

func TestValidateTicket(t *testing.T) {
	// Создаем моки
	keyManager := &MockKeyManager{}
	accessManager := &MockAccessManager{hasAccess: true}

	// Генерируем ключи для тестового сервиса
	publicKey, privateKey, err := keyManager.GenerateKeyPair()
	assert.NoError(t, err)

	repo := &MockServiceRepository{
		service: &Service{
			ID:        1,
			Name:      "test-service",
			PublicKey: EncodeKey(publicKey),
		},
	}

	service := NewTicketService(repo, keyManager, accessManager)

	// Создаем тестовый тикет
	ticket := &Ticket{
		From:     1,
		To:       2,
		TTL:      time.Now().Add(24 * time.Hour).Unix(),
		Metadata: "{}",
	}

	// Подписываем тикет
	data, err := json.Marshal(ticket)
	assert.NoError(t, err)
	signature, err := keyManager.Sign(data, privateKey)
	assert.NoError(t, err)
	ticket.Signature = EncodeKey(signature)

	// Тест: Успешная валидация тикета
	err = service.ValidateTicket(context.Background(), ticket)
	assert.NoError(t, err)

	// Тест: Истекший тикет
	ticket.TTL = time.Now().Add(-1 * time.Hour).Unix()
	err = service.ValidateTicket(context.Background(), ticket)
	assert.Error(t, err)
	assert.Equal(t, ErrTicketExpired, err)

	// Тест: Неверная подпись
	ticket.TTL = time.Now().Add(24 * time.Hour).Unix()
	ticket.Signature = "invalid-signature"
	err = service.ValidateTicket(context.Background(), ticket)
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidSignature, err)
}

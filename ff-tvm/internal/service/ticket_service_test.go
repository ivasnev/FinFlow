package service

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGenerateTicket(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем моки
	keyManager := NewMockKeyManager(ctrl)
	accessManager := NewMockAccessManager(ctrl)
	repo := NewMockServiceRepository(ctrl)

	// Генерируем ключи для тестового сервиса
	_, privateKey, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	data := struct {
		From int   `json:"from"`
		To   int   `json:"to"`
		TTL  int64 `json:"ttl"`
	}{
		From: 1,
		To:   2,
		TTL:  time.Now().Add(24 * time.Hour).Unix(),
	}

	// Сериализуем данные в JSON
	jsonData, _ := json.Marshal(data)

	// Настраиваем ожидания для моков
	accessManager.EXPECT().CheckAccess(1, 2).Return(true)
	repo.EXPECT().GetPrivateKeyHash(gomock.Any(), 1).Return(HashKey(privateKey), nil)
	keyManager.EXPECT().Sign(gomock.Any(), gomock.Any()).Return(ed25519.Sign(privateKey, jsonData), nil)

	svc := NewTicketService(repo, keyManager, accessManager)

	// Тест: Успешная генерация тикета
	ticket, err := svc.GenerateTicket(context.Background(), 1, 2, EncodeKey(privateKey))
	assert.NoError(t, err)
	assert.NotNil(t, ticket)
	assert.Equal(t, 1, ticket.From)
	assert.Equal(t, 2, ticket.To)
	assert.NotEmpty(t, ticket.Signature)
	assert.NotEmpty(t, ticket.Metadata)

	// Тест: Отказ в доступе
	accessManager.EXPECT().CheckAccess(1, 2).Return(false)
	ticket, err = svc.GenerateTicket(context.Background(), 1, 2, "test-service-private-key")
	assert.Error(t, err)
	assert.Equal(t, ErrAccessDenied, err)
}

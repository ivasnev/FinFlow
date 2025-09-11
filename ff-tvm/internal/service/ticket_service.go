package service

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"time"
)

// ticketServiceImpl реализует интерфейс TicketService
type ticketServiceImpl struct {
	repo          ServiceRepository
	keyManager    KeyManager
	accessManager AccessManager
}

// NewTicketService создает новый сервис для работы с тикетами
func NewTicketService(repo ServiceRepository, keyManager KeyManager, accessManager AccessManager) TicketService {
	return &ticketServiceImpl{
		repo:          repo,
		keyManager:    keyManager,
		accessManager: accessManager,
	}
}

// GenerateTicket генерирует новый тикет
func (s *ticketServiceImpl) GenerateTicket(ctx context.Context, from, to int64, secret string) (*Ticket, error) {
	// Проверяем доступ
	if !s.accessManager.CheckAccess(from, to) {
		return nil, ErrAccessDenied
	}

	// Получаем хеш приватного ключа сервиса
	privateKeyHash, err := s.repo.GetPrivateKeyHash(ctx, from)
	if err != nil {
		return nil, err
	}

	// Декодируем секрет из base64
	secretBytes, err := DecodeKey(secret)
	if err != nil {
		return nil, ErrInvalidSecret
	}

	// Проверяем, соответствует ли секрет хешу приватного ключа
	if !VerifyKeyHash(secretBytes, privateKeyHash) {
		return nil, ErrInvalidSecret
	}

	// Создаем данные для подписи
	data := struct {
		From int64 `json:"from"`
		To   int64 `json:"to"`
		TTL  int64 `json:"ttl"`
	}{
		From: from,
		To:   to,
		TTL:  time.Now().Add(24 * time.Hour).Unix(),
	}

	// Сериализуем данные в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	// Подписываем данные
	signature, err := s.keyManager.Sign(jsonData, secretBytes)
	if err != nil {
		return nil, err
	}

	return &Ticket{
		From:      from,
		To:        to,
		TTL:       data.TTL,
		Signature: EncodeKey(signature),
		Metadata:  "{}",
	}, nil
}

// ValidateTicket проверяет валидность тикета
func (s *ticketServiceImpl) ValidateTicket(ctx context.Context, ticket *Ticket) error {
	// Проверяем TTL
	if time.Now().Unix() > ticket.TTL {
		return ErrTicketExpired
	}

	// Получаем публичный ключ сервиса-отправителя
	publicKeyStr, err := s.repo.GetPublicKey(ctx, ticket.From)
	if err != nil {
		return err
	}

	// Декодируем публичный ключ
	publicKey, err := DecodeKey(publicKeyStr)
	if err != nil {
		return err
	}

	// Создаем данные для проверки подписи
	data := struct {
		From int64 `json:"from"`
		To   int64 `json:"to"`
		TTL  int64 `json:"ttl"`
	}{
		From: ticket.From,
		To:   ticket.To,
		TTL:  ticket.TTL,
	}

	// Сериализуем данные в JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Декодируем подпись
	signature, err := DecodeKey(ticket.Signature)
	if err != nil {
		return err
	}

	// Проверяем подпись
	if !s.keyManager.Verify(jsonData, signature, publicKey) {
		return ErrInvalidSignature
	}

	return nil
}

// ValidateTicketSignature проверяет подпись тикета
func ValidateTicketSignature(ticket *Ticket, publicKey string) error {
	// Декодируем публичный ключ
	pubKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return err
	}

	// Декодируем подпись
	signature, err := base64.StdEncoding.DecodeString(ticket.Signature)
	if err != nil {
		return err
	}

	// Подготовка данных для проверки
	data, err := json.Marshal(ticket)
	if err != nil {
		return err
	}

	// Проверяем подпись
	if !ed25519.Verify(ed25519.PublicKey(pubKey), data, signature) {
		return ErrInvalidSignature
	}

	return nil
}

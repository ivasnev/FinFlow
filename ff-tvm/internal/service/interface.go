package service

import (
	"context"
	"crypto/ed25519"
)

//go:generate mockgen -source=interface.go -destination=service_mock.go -package=service

// Service представляет информацию о сервисе
type Service struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	PublicKey      string `json:"public_key"`
	PrivateKeyHash string `json:"-"`
}

type TicketPayload struct {
	From     int    `json:"from"`
	To       int    `json:"to"`
	TTL      int64  `json:"ttl"`
	Metadata string `json:"metadata"`
}

// Ticket представляет тикет авторизации
type Ticket struct {
	Payload   TicketPayload `json:"payload"`
	Signature string        `json:"signature"`
}

// ServiceRepository интерфейс для работы с хранилищем сервисов
type ServiceRepository interface {
	Create(ctx context.Context, service *Service) error
	GetByID(ctx context.Context, id int) (*Service, error)
	GetPublicKey(ctx context.Context, id int) (string, error)
	GetPrivateKeyHash(ctx context.Context, id int) (string, error)
	GrantAccess(ctx context.Context, from, to int) error
	RevokeAccess(ctx context.Context, from, to int) error
}

// KeyManager интерфейс для управления ключами
type KeyManager interface {
	GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error)
	Sign(data []byte, privateKey ed25519.PrivateKey) ([]byte, error)
	Verify(data, signature []byte, publicKey ed25519.PublicKey) bool
}

// AccessManager интерфейс для управления доступом
type AccessManager interface {
	CheckAccess(from, to int) bool
	GrantAccess(from, to int) error
	RevokeAccess(from, to int) error
}

// TicketService интерфейс для работы с тикетами
type TicketService interface {
	GenerateTicket(ctx context.Context, from, to int, secret string) (*Ticket, error)
}

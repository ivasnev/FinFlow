package service

import (
	"context"
	"crypto/ed25519"
)

// Service представляет информацию о сервисе
type Service struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	PublicKey      string `json:"public_key"`
	PrivateKeyHash string `json:"-"`
}

// Ticket представляет тикет авторизации
type Ticket struct {
	From      int64  `json:"from"`
	To        int64  `json:"to"`
	TTL       int64  `json:"ttl"`
	Signature string `json:"signature"`
	Metadata  string `json:"metadata"`
}

// ServiceRepository интерфейс для работы с хранилищем сервисов
type ServiceRepository interface {
	Create(ctx context.Context, service *Service) error
	GetByID(ctx context.Context, id int64) (*Service, error)
	GetPublicKey(ctx context.Context, id int64) (string, error)
	GetPrivateKeyHash(ctx context.Context, id int64) (string, error)
	GrantAccess(ctx context.Context, from, to int64) error
	RevokeAccess(ctx context.Context, from, to int64) error
}

// KeyManager интерфейс для управления ключами
type KeyManager interface {
	GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error)
	Sign(data []byte, privateKey ed25519.PrivateKey) ([]byte, error)
	Verify(data, signature []byte, publicKey ed25519.PublicKey) bool
}

// AccessManager интерфейс для управления доступом
type AccessManager interface {
	CheckAccess(from, to int64) bool
	GrantAccess(from, to int64) error
	RevokeAccess(from, to int64) error
}

// TicketService интерфейс для работы с тикетами
type TicketService interface {
	GenerateTicket(ctx context.Context, from, to int64, secret string) (*Ticket, error)
	ValidateTicket(ctx context.Context, ticket *Ticket) error
}

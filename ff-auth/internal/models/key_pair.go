package models

import (
	"time"
)

// KeyPair представляет пару ключей (публичный и приватный) для подписи токенов
type KeyPair struct {
	ID         int       `json:"id"`
	PublicKey  string    `json:"public_key"`
	PrivateKey string    `json:"-"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

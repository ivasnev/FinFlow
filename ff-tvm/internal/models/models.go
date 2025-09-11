package models

import (
	"time"
)

// Service представляет информацию о сервисе в БД
type Service struct {
	ID        int64  `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	PublicKey string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ServiceAccess представляет разрешения доступа между сервисами
type ServiceAccess struct {
	ID        int64 `gorm:"primaryKey"`
	FromID    int64 `gorm:"not null"`
	ToID      int64 `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// KeyPair представляет пару ключей ED25519
type KeyPair struct {
	ID         int64  `gorm:"primaryKey"`
	ServiceID  int64  `gorm:"not null"`
	PublicKey  string `gorm:"not null"`
	PrivateKey string `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

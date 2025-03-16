package models

import (
	"time"
)

// Service представляет зарегистрированный микросервис
type Service struct {
	ID           uint      `gorm:"primarykey"`
	Name         string    `gorm:"uniqueIndex;not null"`
	Description  string
	PublicKey    string    `gorm:"type:text"`
	PrivateKey   string    `gorm:"type:text"`
	Active       bool      `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastAccessAt *time.Time
}

// ServiceAccess определяет права доступа между сервисами
type ServiceAccess struct {
	ID              uint      `gorm:"primarykey"`
	SourceServiceID uint      `gorm:"not null"`
	TargetServiceID uint      `gorm:"not null"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ExpiresAt       *time.Time
}

// ServiceTicket представляет выданный тикет
type ServiceTicket struct {
	ID              uint      `gorm:"primarykey"`
	SourceServiceID uint      `gorm:"not null"`
	TargetServiceID uint      `gorm:"not null"`
	Token           string    `gorm:"type:text;not null"`
	ExpiresAt       time.Time
	CreatedAt       time.Time
}

// KeyRotation хранит историю ротации ключей
type KeyRotation struct {
	ID        uint      `gorm:"primarykey"`
	ServiceID uint      `gorm:"not null"`
	OldKey    string    `gorm:"type:text"`
	NewKey    string    `gorm:"type:text"`
	CreatedAt time.Time
} 
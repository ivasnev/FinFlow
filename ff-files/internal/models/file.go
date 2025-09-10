package models

import (
	"time"
)

// File представляет метаданные файла
type File struct {
	ID           string    `gorm:"primarykey;type:uuid"`
	Name         string    `gorm:"not null"`
	Path         string    `gorm:"not null"`
	Size         int64     `gorm:"not null"`
	MimeType     string    `gorm:"not null"`
	Hash         string    `gorm:"index"`
	UploadedBy   uint      `gorm:"not null"` // ID пользователя из ID-сервиса
	IsDeleted    bool      `gorm:"default:false"`
	DeletedAt    *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// FileMetadata представляет дополнительные метаданные файла
type FileMetadata struct {
	ID        uint      `gorm:"primarykey"`
	FileID    string    `gorm:"type:uuid;not null"`
	Key       string    `gorm:"not null"`
	Value     string    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TemporaryURL представляет временную ссылку на файл
type TemporaryURL struct {
	ID        string    `gorm:"primarykey;type:uuid"`
	FileID    string    `gorm:"type:uuid;not null"`
	URL       string    `gorm:"not null"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
} 
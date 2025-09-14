package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type File struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Size        int64     `gorm:"not null"`
	MimeType    string    `gorm:"not null"`
	OwnerID     string    `gorm:"not null"`
	UploadedAt  time.Time `gorm:"not null"`
	Metadata    string    `gorm:"type:jsonb"`
	StoragePath string    `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

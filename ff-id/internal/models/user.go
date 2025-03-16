package models

import (
	"time"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID              uint      `gorm:"primarykey"`
	Email           string    `gorm:"uniqueIndex;size:255"`
	Phone           string    `gorm:"uniqueIndex;size:20"`
	Password        string    `gorm:"size:255"`
	Role            Role      `gorm:"type:varchar(20);default:'user'"`
	FirstName       string    `gorm:"size:100"`
	LastName        string    `gorm:"size:100"`
	AvatarID        string    `gorm:"size:36"` // UUID файла из ff-files
	EmailConfirmed  bool      `gorm:"default:false"`
	PhoneConfirmed  bool      `gorm:"default:false"`
	TwoFactorEnabled bool     `gorm:"default:false"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastLoginAt     *time.Time
}

type UserSession struct {
	ID           uint      `gorm:"primarykey"`
	UserID       uint      `gorm:"not null"`
	RefreshToken string    `gorm:"not null"`
	UserAgent    string
	ClientIP     string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type VerificationCode struct {
	ID        uint      `gorm:"primarykey"`
	UserID    uint      `gorm:"not null"`
	Code      string    `gorm:"not null"`
	Type      string    `gorm:"not null"` // email, phone, password_reset
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Role struct {
	ID          uint      `gorm:"primarykey"`
	Name        string    `gorm:"uniqueIndex;size:50"`
	Description string    `gorm:"size:255"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type UserRole struct {
	UserID    uint      `gorm:"primarykey"`
	RoleID    uint      `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
} 
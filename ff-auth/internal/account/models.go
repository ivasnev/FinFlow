package account

import (
	"github.com/google/uuid"
	"time"
)

type Pending struct {
	ID                             uuid.UUID `gorm:"type:uuid;primaryKey"`
	SecretHash                     string
	Email, OriginalEmail, Nickname string
	Phone                          *string
	Name                           *string
	PasswordHash                   string
	UserID                         *int64
	ExistingUser                   bool
	ExpiresAt, CreatedAt           time.Time
}

func (Pending) TableName() string { return "account_pending" }

type EmailToken struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey"`
	TokenHash, Purpose   string
	PendingID            *uuid.UUID
	UserID               *int64
	Email, PasswordHash  string
	ExpiresAt, CreatedAt time.Time
	UsedAt               *time.Time
}

func (EmailToken) TableName() string { return "account_email_tokens" }

type Security struct {
	UserID        int64 `gorm:"primaryKey"`
	VerifiedEmail string
	VerifiedAt    time.Time
}

func (Security) TableName() string { return "account_security" }

type Outbox struct {
	ID                           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Payload                      []byte
	Attempts                     int
	AvailableAt, CreatedAt       time.Time
	LeaseUntil, SentAt, FailedAt *time.Time
	LeaseToken                   *uuid.UUID
}

func (Outbox) TableName() string { return "account_mail_outbox" }

type credentials struct {
	ID                   int64
	PasswordHash         string
	Email, Nickname      string
	CreatedAt, UpdatedAt time.Time
}

func (credentials) TableName() string { return "users" }

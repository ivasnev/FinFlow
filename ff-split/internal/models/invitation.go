package models

import "time"

// Invitation хранит только хеш случайного секрета и ограниченное время доступа.
type Invitation struct {
	ID        int64
	EventID   int64
	TokenHash string
	CreatedBy int64
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// TableName фиксирует таблицу приглашений.
func (Invitation) TableName() string { return "event_invitations" }

// InvitationPreview содержит только название события до явного вступления.
type InvitationPreview struct {
	EventID   int64
	EventName string
	ExpiresAt time.Time
	IsMember  bool
}

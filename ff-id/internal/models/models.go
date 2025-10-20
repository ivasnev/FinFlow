package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя системы (доменная модель)
type User struct {
	ID        int64
	Email     string
	Phone     sql.NullString
	Nickname  string
	Name      sql.NullString
	Birthdate sql.NullTime
	AvatarID  uuid.NullUUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserAvatar представляет аватарку пользователя (доменная модель)
type UserAvatar struct {
	ID         uuid.UUID
	UserID     int64
	FileID     uuid.UUID
	UploadedAt time.Time
}

// UserFriend представляет связь дружбы между пользователями (доменная модель)
type UserFriend struct {
	ID        int64
	UserID    int64
	FriendID  int64
	Status    string
	CreatedAt time.Time
	User      User
	Friend    User
}

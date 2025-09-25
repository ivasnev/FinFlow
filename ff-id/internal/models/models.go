package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User представляет пользователя системы
type User struct {
	ID        int64          `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Email     string         `gorm:"type:text;unique;not null;column:email" json:"email"`
	Phone     sql.NullString `gorm:"type:text;unique;column:phone" json:"phone,omitempty"`
	Nickname  string         `gorm:"type:text;unique;not null;column:nickname" json:"nickname"`
	Name      sql.NullString `gorm:"type:text;column:name" json:"name,omitempty"`
	Birthdate sql.NullTime   `gorm:"type:date;column:birthdate" json:"birthdate,omitempty"`
	AvatarID  uuid.NullUUID  `gorm:"type:uuid;column:avatar" json:"avatar_id,omitempty"`
	CreatedAt time.Time      `gorm:"type:timestamp;not null;default:now();column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamp;not null;default:now();column:updated_at" json:"updated_at"`

	// Связи
	Avatars []UserAvatar `gorm:"foreignKey:UserID" json:"avatars,omitempty"`
	Friends []UserFriend `gorm:"foreignKey:UserID" json:"-"`
}

// TableName устанавливает имя таблицы для модели User
func (User) TableName() string {
	return "users"
}

// BeforeUpdate обновляет поле updated_at перед сохранением изменений
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

// UserAvatar представляет аватарку пользователя
type UserAvatar struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	UserID     int64     `gorm:"type:bigint;not null;column:user_id" json:"user_id"`
	FileID     uuid.UUID `gorm:"type:uuid;not null;column:file_id" json:"file_id"`
	UploadedAt time.Time `gorm:"type:timestamp;not null;default:now();column:uploaded_at" json:"uploaded_at"`
}

// TableName устанавливает имя таблицы для модели UserAvatar
func (UserAvatar) TableName() string {
	return "user_avatars"
}

// UserFriend представляет связь дружбы между пользователями
type UserFriend struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64     `gorm:"type:bigint;not null;column:user_id;index:idx_user_friend" json:"user_id"`
	FriendID  int64     `gorm:"type:bigint;not null;column:friend_id;index:idx_user_friend" json:"friend_id"`
	Status    string    `gorm:"type:varchar(20);not null;column:status;default:'pending'" json:"status"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now();column:created_at" json:"created_at"`

	// Связи
	User   User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Friend User `gorm:"foreignKey:FriendID" json:"friend,omitempty"`
}

// TableName устанавливает имя таблицы для модели UserFriend
func (UserFriend) TableName() string {
	return "user_friends"
}

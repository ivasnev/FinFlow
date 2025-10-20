package friend

import (
	"time"

	"gorm.io/gorm"
)

// UserFriend представляет модель связи дружбы в базе данных
type UserFriend struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id"`
	UserID    int64     `gorm:"type:bigint;not null;column:user_id;index:idx_user_friend"`
	FriendID  int64     `gorm:"type:bigint;not null;column:friend_id;index:idx_user_friend"`
	Status    string    `gorm:"type:varchar(20);not null;column:status;default:'pending'"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now();column:created_at"`
}

// TableName устанавливает имя таблицы для модели UserFriend
func (UserFriend) TableName() string {
	return "user_friends"
}

// User представляет модель пользователя для связей (вложенная структура)
type User struct {
	ID       int64  `gorm:"primaryKey;autoIncrement;column:id"`
	Email    string `gorm:"type:text;unique;not null;column:email"`
	Nickname string `gorm:"type:text;unique;not null;column:nickname"`
	Name     string `gorm:"type:text;column:name"`
	AvatarID string `gorm:"type:uuid;column:avatar"`
}

// TableName устанавливает имя таблицы для модели User
func (User) TableName() string {
	return "users"
}

// BeforeCreate устанавливает значения по умолчанию
func (uf *UserFriend) BeforeCreate(tx *gorm.DB) error {
	if uf.CreatedAt.IsZero() {
		uf.CreatedAt = time.Now()
	}
	return nil
}

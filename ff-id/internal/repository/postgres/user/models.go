package user

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User представляет модель пользователя в базе данных
type User struct {
	ID        int64          `gorm:"primaryKey;autoIncrement;column:id"`
	Email     string         `gorm:"type:text;unique;not null;column:email"`
	Phone     sql.NullString `gorm:"type:text;unique;column:phone"`
	Nickname  string         `gorm:"type:text;unique;not null;column:nickname"`
	Name      sql.NullString `gorm:"type:text;column:name"`
	Birthdate sql.NullTime   `gorm:"type:date;column:birthdate"`
	AvatarID  uuid.NullUUID  `gorm:"type:uuid;column:avatar"`
	CreatedAt time.Time      `gorm:"type:timestamp;not null;default:now();column:created_at"`
	UpdatedAt time.Time      `gorm:"type:timestamp;not null;default:now();column:updated_at"`
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

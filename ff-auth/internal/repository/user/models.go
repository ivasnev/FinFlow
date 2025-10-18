package user

import (
	"time"

	"gorm.io/gorm"
)

// User представляет пользователя системы
type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Email        string    `gorm:"type:text;unique;not null;column:email" json:"email"`
	PasswordHash string    `gorm:"type:text;not null;column:password_hash" json:"-"`
	Nickname     string    `gorm:"type:text;unique;not null;column:nickname" json:"nickname"`
	CreatedAt    time.Time `gorm:"type:timestamp;not null;default:now();column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;not null;default:now();column:updated_at" json:"updated_at"`
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

// RoleEntity представляет роль в системе
type RoleEntity struct {
	ID   int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name string `gorm:"type:text;unique;not null;column:name" json:"name"`
}

// TableName устанавливает имя таблицы для модели RoleEntity
func (RoleEntity) TableName() string {
	return "roles"
}

// UserRole представляет связь между пользователем и ролью
type UserRole struct {
	UserID int64 `gorm:"primaryKey;column:user_id" json:"user_id"`
	RoleID int   `gorm:"primaryKey;column:role_id" json:"role_id"`
}

// TableName устанавливает имя таблицы для модели UserRole
func (UserRole) TableName() string {
	return "user_roles"
}

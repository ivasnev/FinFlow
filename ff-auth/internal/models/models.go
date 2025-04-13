package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Role определяет роль пользователя в системе
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
)

// User представляет пользователя системы
type User struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Email        string    `gorm:"type:text;unique;not null;column:email" json:"email"`
	PasswordHash string    `gorm:"type:text;not null;column:password_hash" json:"-"`
	Nickname     string    `gorm:"type:text;unique;not null;column:nickname" json:"nickname"`
	CreatedAt    time.Time `gorm:"type:timestamp;not null;default:now();column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;not null;default:now();column:updated_at" json:"updated_at"`

	// Связи
	Roles        []UserRole     `gorm:"foreignKey:UserID" json:"roles,omitempty"`
	Sessions     []Session      `gorm:"foreignKey:UserID" json:"sessions,omitempty"`
	LoginHistory []LoginHistory `gorm:"foreignKey:UserID" json:"login_history,omitempty"`
	Devices      []Device       `gorm:"foreignKey:UserID" json:"devices,omitempty"`
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

// Session представляет активную сессию пользователя
type Session struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	UserID       int64          `gorm:"type:bigint;not null;column:user_id" json:"user_id"`
	RefreshToken string         `gorm:"type:text;unique;not null;column:refresh_token" json:"-"`
	IPAddress    pq.StringArray `gorm:"type:inet;column:ip_address" json:"ip_address"`
	ExpiresAt    time.Time      `gorm:"type:timestamp;not null;column:expires_at" json:"expires_at"`
	CreatedAt    time.Time      `gorm:"type:timestamp;not null;default:now();column:created_at" json:"created_at"`
}

// TableName устанавливает имя таблицы для модели Session
func (Session) TableName() string {
	return "sessions"
}

// LoginHistory представляет историю входов пользователя
type LoginHistory struct {
	ID        int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64     `gorm:"type:bigint;not null;column:user_id" json:"user_id"`
	IPAddress string    `gorm:"type:inet;not null;column:ip_address" json:"ip_address"`
	UserAgent string    `gorm:"type:text;column:user_agent" json:"user_agent"`
	CreatedAt time.Time `gorm:"type:timestamp;not null;default:now();column:created_at" json:"created_at"`
}

// TableName устанавливает имя таблицы для модели LoginHistory
func (LoginHistory) TableName() string {
	return "login_history"
}

// Device представляет устройство, с которого пользователь входил в систему
type Device struct {
	ID        int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	UserID    int64     `gorm:"type:bigint;not null;column:user_id" json:"user_id"`
	DeviceID  string    `gorm:"type:text;unique;not null;column:device_id" json:"device_id"`
	UserAgent string    `gorm:"type:text;not null;column:user_agent" json:"user_agent"`
	LastLogin time.Time `gorm:"type:timestamp;not null;default:now();column:last_login" json:"last_login"`
}

// TableName устанавливает имя таблицы для модели Device
func (Device) TableName() string {
	return "devices"
}

// KeyPair представляет пару ключей (публичный и приватный) для подписи токенов
type KeyPair struct {
	ID         int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	PublicKey  string    `gorm:"type:text;not null;column:public_key" json:"public_key"`
	PrivateKey string    `gorm:"type:text;not null;column:private_key" json:"-"`
	IsActive   bool      `gorm:"type:boolean;not null;default:true;column:is_active" json:"is_active"`
	CreatedAt  time.Time `gorm:"type:timestamp;not null;default:now();column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"type:timestamp;not null;default:now();column:updated_at" json:"updated_at"`
}

// TableName устанавливает имя таблицы для модели KeyPair
func (KeyPair) TableName() string {
	return "key_pairs"
}

// BeforeUpdate обновляет поле updated_at перед сохранением изменений
func (k *KeyPair) BeforeUpdate(tx *gorm.DB) error {
	k.UpdatedAt = time.Now()
	return nil
}

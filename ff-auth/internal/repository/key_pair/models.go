package key_pair

import (
	"time"

	"gorm.io/gorm"
)

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

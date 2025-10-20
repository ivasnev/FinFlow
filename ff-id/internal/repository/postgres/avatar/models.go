package avatar

import (
	"time"

	"github.com/google/uuid"
)

// UserAvatar представляет модель аватара пользователя в базе данных
type UserAvatar struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;column:id"`
	UserID     int64     `gorm:"type:bigint;not null;column:user_id"`
	FileID     uuid.UUID `gorm:"type:uuid;not null;column:file_id"`
	UploadedAt time.Time `gorm:"type:timestamp;not null;default:now();column:uploaded_at"`
}

// TableName устанавливает имя таблицы для модели UserAvatar
func (UserAvatar) TableName() string {
	return "user_avatars"
}

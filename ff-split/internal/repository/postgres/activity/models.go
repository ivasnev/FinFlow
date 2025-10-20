package activity

import "time"

// Activity представляет действие в системе в БД
type Activity struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement"`
	EventID     *int64    `gorm:"column:event_id"`
	UserID      *int64    `gorm:"column:user_id"`
	Description string    `gorm:"column:description"`
	IconID      int       `gorm:"column:icon_id"`
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
}

// TableName задает имя таблицы для модели Activity
func (Activity) TableName() string {
	return "activities"
}

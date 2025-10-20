package task

import "time"

// Task представляет задачу в системе в БД
type Task struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement"`
	UserID      *int64    `gorm:"column:user_id"`
	EventID     *int64    `gorm:"column:event_id"`
	Title       string    `gorm:"column:title;not null"`
	Description string    `gorm:"column:description"`
	Priority    int       `gorm:"column:priority;default:0"`
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
}

// TableName задает имя таблицы для модели Task
func (Task) TableName() string {
	return "tasks"
}

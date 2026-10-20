package access

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// Repository проверяет участие и привязку объектов к мероприятию.
type Repository struct{ db *gorm.DB }

// NewRepository создаёт репозиторий проверки доступа.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// IsMember сопоставляет внешний ID из токена с внутренней записью участника.
func (r *Repository) IsMember(ctx context.Context, externalID, eventID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("user_event AS ue").
		Joins("JOIN users AS u ON u.id = ue.user_id").
		Where("u.user_id = ? AND ue.event_id = ? AND u.is_dummy = false", externalID, eventID).
		Count(&count).Error
	return count > 0, err
}

// HasResource проверяет объект одновременно по его ID и мероприятию.
func (r *Repository) HasResource(ctx context.Context, eventID int64, kind string, resourceID int64) (bool, error) {
	var table string
	switch kind {
	case "settlement":
		table = "settlements"
	case "transaction":
		table = "transactions"
	case "task":
		table = "tasks"
	case "activity":
		table = "activities"
	case "user":
		table = "user_event"
	default:
		return false, fmt.Errorf("unsupported event resource: %s", kind)
	}
	column := "id"
	if kind == "user" {
		column = "user_id"
	}
	var count int64
	err := r.db.WithContext(ctx).Table(table).Where(column+" = ? AND event_id = ?", resourceID, eventID).Count(&count).Error
	return count > 0, err
}

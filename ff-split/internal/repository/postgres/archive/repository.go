package archive

import (
	"context"
	"time"

	txcontext "github.com/ivasnev/FinFlow/ff-split/internal/common/db"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/history"
	"gorm.io/gorm"
)

// Repository удерживает данные архива и историю в общей транзакции.
type Repository struct{ db *gorm.DB }

// eventRow не содержит ORM-отношений и безопасен для ограниченного SQL-снимка.
type eventRow struct {
	ID          int64
	Name        string
	Description string
	CategoryID  *int
	ImageID     string
	Status      string
	DeletedAt   *time.Time
	PurgeAt     *time.Time
}

// eventModel переводит SQL-снимок в доменную модель.
func eventModel(row eventRow) models.Event {
	return models.Event{ID: row.ID, Name: row.Name, Description: row.Description, CategoryID: row.CategoryID, ImageID: row.ImageID, Status: row.Status, DeletedAt: row.DeletedAt, PurgeAt: row.PurgeAt}
}

// NewRepository создаёт хранилище архива.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// conn возвращает активное соединение транзакции с отменой по контексту.
func (r *Repository) conn(ctx context.Context) *gorm.DB {
	return txcontext.GetTx(ctx, r.db).WithContext(ctx)
}

// List не раскрывает архив постороннего пользователя.
func (r *Repository) List(ctx context.Context, actor int64) ([]models.Event, error) {
	rows := []eventRow{}
	err := r.conn(ctx).Raw("SELECT e.*,event_purge_at(e.deleted_at) AS purge_at FROM events e JOIN user_event ue ON ue.event_id=e.id JOIN users u ON u.id=ue.user_id WHERE u.user_id=? AND NOT u.is_dummy AND e.deleted_at IS NOT NULL ORDER BY e.deleted_at DESC,e.id DESC", actor).Scan(&rows).Error
	result := make([]models.Event, 0, len(rows))
	for _, row := range rows {
		result = append(result, eventModel(row))
	}
	return result, err
}

// InEvent синхронизирует архивирование, восстановление и окончательную очистку одной строкой события.
func (r *Repository) InEvent(ctx context.Context, id int64, fn func(context.Context, *models.Event) error) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return txcontext.WithTx(ctx, r.db, func(ctx context.Context) error {
		var row eventRow
		result := r.conn(ctx).Raw("SELECT e.*,event_purge_at(deleted_at) AS purge_at FROM events e WHERE id=? FOR UPDATE", id).Scan(&row)
		if result.Error != nil {
			return result.Error
		}
		if row.ID == 0 {
			return gorm.ErrRecordNotFound
		}
		event := eventModel(row)
		return fn(ctx, &event)
	})
}

// IsMember проверяет реальное участие внутри транзакции.
func (r *Repository) IsMember(ctx context.Context, eventID, actor int64) (bool, error) {
	var count int64
	err := r.conn(ctx).Raw("SELECT COUNT(*) FROM user_event ue JOIN users u ON u.id=ue.user_id WHERE ue.event_id=? AND u.user_id=? AND NOT u.is_dummy", eventID, actor).Scan(&count).Error
	return count > 0, err
}

// SetArchived сохраняет срок и отзывает приглашения; восстановление не оживляет старые ссылки.
func (r *Repository) SetArchived(ctx context.Context, row *models.Event, archived bool) error {
	tx := r.conn(ctx)
	before := map[string]string{"Статус": "Активно"}
	after := map[string]string{"Статус": "В архиве"}
	kind := "event.archived"
	if archived {
		var dates struct {
			DeletedAt time.Time
			PurgeAt   time.Time
		}
		if err := tx.Raw("UPDATE events SET deleted_at=CURRENT_TIMESTAMP,status='archive' WHERE id=? RETURNING deleted_at,event_purge_at(deleted_at) AS purge_at", row.ID).Scan(&dates).Error; err != nil {
			return err
		}
		after["Удаление после"] = dates.PurgeAt.Format(time.RFC3339)
		var links []struct{ ID int64 }
		if err := tx.Raw("UPDATE event_invitations SET revoked_at=CURRENT_TIMESTAMP WHERE event_id=? AND revoked_at IS NULL RETURNING id", row.ID).Scan(&links).Error; err != nil {
			return err
		}
		for _, link := range links {
			if err := history.Record(ctx, tx, row.ID, "invitation.revoked", link.ID, map[string]string{"Приглашение": "Действует"}, map[string]string{"Приглашение": "Событие в архиве"}); err != nil {
				return err
			}
		}
	} else {
		kind = "event.restored"
		before, after = after, before
		if err := tx.Exec("UPDATE events SET deleted_at=NULL,status='active' WHERE id=?", row.ID).Error; err != nil {
			return err
		}
	}
	return history.Record(ctx, tx, row.ID, kind, row.ID, before, after)
}

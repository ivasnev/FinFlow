package invitation

import (
	"context"
	"time"

	txcontext "github.com/ivasnev/FinFlow/ff-split/internal/common/db"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/history"
	"gorm.io/gorm"
)

// Repository хранит приглашения; изменение ссылки и вступление сериализованы событием.
type Repository struct{ db *gorm.DB }

// NewRepository создаёт хранилище приглашений.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// conn возвращает активную транзакцию с контекстом запроса.
func (r *Repository) conn(ctx context.Context) *gorm.DB {
	return txcontext.GetTx(ctx, r.db).WithContext(ctx)
}

// InEvent ограничивает ожидание и держит событие до commit.
func (r *Repository) InEvent(ctx context.Context, id int64, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return txcontext.WithTx(ctx, r.db, func(ctx context.Context) error {
		tx := r.conn(ctx)
		var found int64
		if err := tx.Raw("SELECT id FROM events WHERE id=? FOR UPDATE", id).Scan(&found).Error; err != nil {
			return err
		}
		if found == 0 {
			return gorm.ErrRecordNotFound
		}
		return fn(ctx)
	})
}

// IsMember сопоставляет внешний ID только реальному участнику.
func (r *Repository) IsMember(ctx context.Context, eventID, externalID int64) (bool, error) {
	var count int64
	err := r.conn(ctx).Raw("SELECT COUNT(*) FROM user_event ue JOIN users u ON u.id=ue.user_id WHERE ue.event_id=? AND u.user_id=? AND NOT u.is_dummy", eventID, externalID).Scan(&count).Error
	return count > 0, err
}

// List возвращает только собственные действующие ссылки, без секретов.
func (r *Repository) List(ctx context.Context, eventID, actor int64) ([]models.Invitation, error) {
	rows := []models.Invitation{}
	err := r.conn(ctx).Where("event_id=? AND created_by=? AND revoked_at IS NULL AND expires_at>clock_timestamp()", eventID, actor).Order("id DESC").Limit(5).Find(&rows).Error
	return rows, err
}

// Create вставляет хеш; исходный секрет не попадает ни в таблицу, ни в историю.
func (r *Repository) Create(ctx context.Context, row *models.Invitation) error {
	tx := r.conn(ctx)
	if err := tx.Raw("INSERT INTO event_invitations(event_id,token_hash,created_by) VALUES(?,?,?) RETURNING id,event_id,token_hash,created_by,created_at,expires_at,revoked_at", row.EventID, row.TokenHash, row.CreatedBy).Scan(row).Error; err != nil {
		return err
	}
	return history.Record(ctx, tx, row.EventID, "invitation.created", row.ID, nil, map[string]string{"Действует до": row.ExpiresAt.UTC().Format(time.RFC3339)})
}

// Get ограничивает ссылку её событием.
func (r *Repository) Get(ctx context.Context, eventID, id int64) (*models.Invitation, error) {
	var row models.Invitation
	err := r.conn(ctx).Where("event_id=? AND id=?", eventID, id).Take(&row).Error
	return &row, err
}

// Lookup проверяет срок по часам PostgreSQL при каждом использовании.
func (r *Repository) Lookup(ctx context.Context, hash string) (*models.Invitation, error) {
	var row models.Invitation
	err := r.conn(ctx).Where("token_hash=? AND revoked_at IS NULL AND expires_at>clock_timestamp() AND EXISTS (SELECT 1 FROM user_event ue JOIN users u ON u.id=ue.user_id WHERE ue.event_id=event_invitations.event_id AND u.user_id=event_invitations.created_by AND NOT u.is_dummy)", hash).Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.ErrInvitationUnavailable
	}
	return &row, err
}

// EventName отдаёт только заголовок события до принятия ссылки.
func (r *Repository) EventName(ctx context.Context, id int64) (string, error) {
	var name string
	err := r.conn(ctx).Raw("SELECT name FROM events WHERE id=?", id).Scan(&name).Error
	return name, err
}

// Revoke сохраняет отзыв один раз и не создаёт дубликат истории при повторе.
func (r *Repository) Revoke(ctx context.Context, row *models.Invitation) error {
	tx := r.conn(ctx)
	result := tx.Model(&models.Invitation{}).Where("id=? AND revoked_at IS NULL", row.ID).UpdateColumn("revoked_at", gorm.Expr("CURRENT_TIMESTAMP"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return nil
	}
	return history.Record(ctx, tx, row.EventID, "invitation.revoked", row.ID, map[string]string{"Приглашение": "Действует"}, map[string]string{"Приглашение": "Отозвано"})
}

// Join атомарно фиксирует только новое участие, оставляя повтор безопасным.
func (r *Repository) Join(ctx context.Context, eventID, userID int64) error {
	tx := r.conn(ctx)
	result := tx.Exec("INSERT INTO user_event(user_id,event_id) VALUES(?,?) ON CONFLICT DO NOTHING", userID, eventID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return nil
	}
	return history.Member(ctx, tx, eventID, userID, true)
}

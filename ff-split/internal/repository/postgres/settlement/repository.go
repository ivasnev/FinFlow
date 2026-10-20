package settlement

import (
	"context"
	"errors"
	"time"

	txcontext "github.com/ivasnev/FinFlow/ff-split/internal/common/db"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/history"
	"gorm.io/gorm"
)

// Repository хранит возвраты на одном соединении транзакции.
type Repository struct{ db *gorm.DB }

// NewRepository создаёт репозиторий журнала возвратов.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// conn возвращает текущую транзакцию, уважая отмену операции.
func (r *Repository) conn(ctx context.Context) *gorm.DB {
	return txcontext.GetTx(ctx, r.db).WithContext(ctx)
}

// InEvent блокирует только изменения конкретного мероприятия до завершения callback.
func (r *Repository) InEvent(ctx context.Context, eventID int64, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := history.LockEvent(tx, eventID); err != nil {
			return err
		}
		return fn(context.WithValue(ctx, txcontext.TxContextKey{}, tx))
	})
}

// Members читает актуальное участие и сопоставляет пространство ID.
func (r *Repository) Members(ctx context.Context, eventID int64) (map[int64]*int64, error) {
	var rows []struct {
		ID     int64
		UserID *int64
	}
	err := r.conn(ctx).Raw("SELECT u.id,CASE WHEN u.is_dummy THEN NULL ELSE u.user_id END AS user_id FROM users u JOIN user_event ue ON ue.user_id=u.id WHERE ue.event_id=? FOR KEY SHARE OF ue", eventID).Scan(&rows).Error
	result := map[int64]*int64{}
	for _, row := range rows {
		result[row.ID] = row.UserID
	}
	return result, err
}

// Ledger читает расходы и возвраты одним SQL-запросом, включая события без оставшихся расходов.
func (r *Repository) Ledger(ctx context.Context, eventID int64) ([]models.Debt, error) {
	var rows []struct {
		FromUserID int64
		ToUserID   int64
		Amount     float64
	}
	err := r.conn(ctx).Raw(`SELECT d.from_user_id,d.to_user_id,d.amount FROM debts d JOIN transactions t ON t.id=d.transaction_id WHERE t.event_id=?
 UNION ALL SELECT to_user_id,from_user_id,amount_cents::numeric/100 AS amount FROM settlements WHERE event_id=? AND canceled_at IS NULL`, eventID, eventID).Scan(&rows).Error
	result := make([]models.Debt, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.Debt{FromUserID: row.FromUserID, ToUserID: row.ToUserID, Amount: row.Amount})
	}
	return result, err
}

// ByRequest читает прежний результат независимо от последующей отмены.
func (r *Repository) ByRequest(ctx context.Context, eventID, actorID int64, key string) (*models.Settlement, error) {
	var row models.Settlement
	err := r.conn(ctx).Where("event_id=? AND created_by=? AND request_id=?", eventID, actorID, key).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &row, err
}

// Get запрещает чтение возврата из другого мероприятия.
func (r *Repository) Get(ctx context.Context, eventID, id int64) (*models.Settlement, error) {
	var row models.Settlement
	err := r.conn(ctx).Where("event_id=? AND id=?", eventID, id).Take(&row).Error
	return &row, err
}

// List возвращает страницу от новых записей к старым без смещения при параллельных вставках.
func (r *Repository) List(ctx context.Context, eventID, before int64) ([]models.Settlement, error) {
	rows := []models.Settlement{}
	query := r.conn(ctx).Where("event_id=?", eventID)
	if before > 0 {
		query = query.Where("id < ?", before)
	}
	err := query.Order("id DESC").Limit(50).Find(&rows).Error
	return rows, err
}

// Create сохраняет запись и только увеличивает версию, без запуска оптимизации.
func (r *Repository) Create(ctx context.Context, row *models.Settlement) error {
	if err := r.conn(ctx).Create(row).Error; err != nil {
		return err
	}
	tx := r.conn(ctx)
	if err := tx.Exec("UPDATE events SET debt_revision=debt_revision+1 WHERE id=?", row.EventID).Error; err != nil {
		return err
	}
	after, err := history.Payment(tx, row)
	if err != nil {
		return err
	}
	return history.Record(ctx, tx, row.EventID, "settlement.created", row.ID, nil, after)
}

// Cancel сохраняет сам факт ошибочной записи и увеличивает версию только при первой отмене.
func (r *Repository) Cancel(ctx context.Context, eventID, id int64) error {
	result := r.conn(ctx).Model(&models.Settlement{}).Where("event_id=? AND id=? AND canceled_at IS NULL", eventID, id).UpdateColumn("canceled_at", gorm.Expr("CURRENT_TIMESTAMP"))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return nil
	}
	tx := r.conn(ctx)
	if err := tx.Exec("UPDATE events SET debt_revision=debt_revision+1 WHERE id=?", eventID).Error; err != nil {
		return err
	}
	row, err := r.Get(ctx, eventID, id)
	if err != nil {
		return err
	}
	before, err := history.Payment(tx, row)
	if err != nil {
		return err
	}
	return history.Record(ctx, tx, eventID, "settlement.canceled", id, before, nil)
}

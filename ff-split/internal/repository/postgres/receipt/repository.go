package receipt

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	dbctx "github.com/ivasnev/FinFlow/ff-split/internal/common/db"
	commonerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/history"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Repository хранит JSON-снимки без зависимостей от HTTP и OCR SDK.
type Repository struct{ db *gorm.DB }

// NewRepository подключает базу без логирования содержимого чеков и ежесекундных опросов.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db.Session(&gorm.Session{Logger: db.Logger.LogMode(logger.Silent)})}
}

// row отделяет JSONB и скрытые поля очереди от публичной модели.
type row struct {
	ID             string
	EventID        int64
	CreatedBy      int64
	Sequence       int64
	Status         string
	Revision       int
	ImageKey       string
	ImageSha256    string
	ContentType    string
	LeaseToken     *string
	Attempts       int
	ImageAvailable bool
	RetryRevision  int
	QueuedAt       time.Time
	Result         string
	Draft          string
	TransactionIDs string
	Error          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// decode проверяет снимки до публикации пользователю.
func decode(r row) (*models.Receipt, error) {
	v := &models.Receipt{ID: r.ID, EventID: r.EventID, CreatedBy: r.CreatedBy, Sequence: r.Sequence, Status: r.Status, Revision: r.Revision, ImageKey: r.ImageKey, ImageSHA256: r.ImageSha256, ContentType: r.ContentType, Attempts: r.Attempts, ImageAvailable: r.ImageAvailable, RetryRevision: r.RetryRevision, QueuedAt: r.QueuedAt, Error: r.Error, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
	if r.LeaseToken != nil {
		v.LeaseToken = *r.LeaseToken
	}
	if err := json.Unmarshal([]byte(r.Result), &v.Result); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(r.Draft), &v.Draft); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(r.TransactionIDs), &v.TransactionIDs); err != nil {
		return nil, err
	}
	if v.Result.Items == nil {
		v.Result.Items = []models.ReceiptItem{}
	}
	if v.Result.Warnings == nil {
		v.Result.Warnings = []string{}
	}
	if v.Draft.Items == nil {
		v.Draft.Items = []models.ReceiptDraftItem{}
	}
	if v.TransactionIDs == nil {
		v.TransactionIDs = []int{}
	}
	return v, nil
}

// InEvent сохраняет единый порядок: событие, затем чек, затем расходы.
func (r *Repository) InEvent(ctx context.Context, eventID int64, fn func(context.Context) error) error {
	return dbctx.WithTx(ctx, r.db, func(ctx context.Context) error {
		tx := dbctx.GetTx(ctx, r.db).WithContext(ctx)
		if err := tx.Exec("SET LOCAL lock_timeout='5s'").Error; err != nil {
			return err
		}
		if err := history.LockEvent(tx, eventID); err != nil {
			return err
		}
		return fn(ctx)
	})
}

// IsMember всегда сопоставляет внешний ID из доверенного контекста.
func (r *Repository) IsMember(ctx context.Context, eventID, actorID int64) (bool, error) {
	var n int64
	err := dbctx.GetTx(ctx, r.db).WithContext(ctx).Raw("SELECT count(*) FROM user_event ue JOIN users u ON u.id=ue.user_id WHERE ue.event_id=? AND u.user_id=? AND NOT u.is_dummy", eventID, actorID).Scan(&n).Error
	return n > 0, err
}

// Create проверяет дедупликацию до квот; лимиты сериализуются только на короткое время записи.
func (r *Repository) Create(ctx context.Context, v *models.Receipt) (*models.Receipt, error) {
	tx := dbctx.GetTx(ctx, r.db).WithContext(ctx)
	var existing row
	if err := tx.Raw("SELECT * FROM receipts WHERE event_id=? AND image_sha256=? AND deleted_at IS NULL FOR UPDATE", v.EventID, v.ImageSHA256).Scan(&existing).Error; err != nil {
		return nil, err
	}
	if existing.ID != "" {
		// Явная повторная загрузка восстанавливает и старые failed-записи:
		// до появления image_available факт успешной записи фото не сохранялся.
		if existing.Status == "failed" {
			if err := r.admit(ctx, v.CreatedBy); err != nil {
				return nil, err
			}
			if err := tx.Exec("INSERT INTO receipt_image_gc(image_key) VALUES(?) ON CONFLICT DO NOTHING", existing.ImageKey).Error; err != nil {
				return nil, err
			}
			var resumed row
			if err := tx.Raw("UPDATE receipts SET image_key=?,image_available=false,status='uploading',queued_by=?,queued_at=clock_timestamp(),attempts=0,error='',lease_token=?,lease_until=clock_timestamp()+interval '60 seconds',revision=revision+1,updated_at=clock_timestamp() WHERE id=? RETURNING *", v.ImageKey, v.CreatedBy, uuid.NewString(), existing.ID).Scan(&resumed).Error; err != nil {
				return nil, err
			}
			if err := tx.Exec("INSERT INTO receipt_admissions(receipt_id,actor_id) VALUES(?,?)", existing.ID, v.CreatedBy).Error; err != nil {
				return nil, err
			}
			return decode(resumed)
		}
		return decode(existing)
	}
	if err := r.admit(ctx, v.CreatedBy); err != nil {
		return nil, err
	}
	var created row
	if err := tx.Raw("INSERT INTO receipts(id,event_id,created_by,image_key,image_sha256,content_type,status,lease_token,lease_until) VALUES(?,?,?,?,?,?,'uploading',?,clock_timestamp()+interval '60 seconds') RETURNING *", v.ID, v.EventID, v.CreatedBy, v.ImageKey, v.ImageSHA256, v.ContentType, uuid.NewString()).Scan(&created).Error; err != nil {
		return nil, err
	}
	if err := tx.Exec("DELETE FROM receipt_image_gc WHERE image_key=?", v.ImageKey).Error; err != nil {
		return nil, err
	}
	return decode(created)
}

// Get связывает UUID с событием и при изменении удерживает строку до commit.
func (r *Repository) Get(ctx context.Context, eventID int64, id string, lock bool) (*models.Receipt, error) {
	var value row
	q := "SELECT * FROM receipts WHERE event_id=? AND id=? AND deleted_at IS NULL"
	if lock {
		q += " FOR UPDATE"
	}
	if err := dbctx.GetTx(ctx, r.db).WithContext(ctx).Raw(q, eventID, id).Scan(&value).Error; err != nil {
		return nil, err
	}
	if value.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return decode(value)
}

// List отдаёт ограниченную страницу без изображений и скрытых lease-полей.
func (r *Repository) List(ctx context.Context, eventID, before int64) (*models.ReceiptList, error) {
	q := r.db.WithContext(ctx).Table("receipts").Where("event_id=? AND deleted_at IS NULL", eventID)
	if before > 0 {
		q = q.Where("sequence<?", before)
	}
	var rows []row
	if err := q.Order("sequence DESC").Limit(21).Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := &models.ReceiptList{Receipts: []models.Receipt{}}
	if len(rows) > 20 {
		rows = rows[:20]
		n := rows[19].Sequence
		result.NextBefore = &n
	}
	for _, row := range rows {
		v, err := decode(row)
		if err != nil {
			return nil, err
		}
		result.Receipts = append(result.Receipts, *v)
	}
	return result, nil
}

// Save проверяет ревизию даже при повторном вызове из другого процесса API.
func (r *Repository) Save(ctx context.Context, v *models.Receipt) error {
	draft, err := json.Marshal(v.Draft)
	if err != nil {
		return err
	}
	ids, err := json.Marshal(v.TransactionIDs)
	if err != nil {
		return err
	}
	result := dbctx.GetTx(ctx, r.db).WithContext(ctx).Exec("UPDATE receipts SET draft=?::jsonb,transaction_ids=?::jsonb,status=?,revision=revision+1,updated_at=clock_timestamp() WHERE id=? AND revision=? AND deleted_at IS NULL", string(draft), string(ids), v.Status, v.ID, v.Revision)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return commonerrors.NewValidationError("receipt", "чек уже изменён; обновите его")
	}
	v.Revision++
	return nil
}

// Claim атомарно арендует одну задачу, не удерживая транзакцию во время OCR.
func (r *Repository) Claim(ctx context.Context) (*models.Receipt, error) {
	var value row
	token := uuid.NewString()
	err := r.db.WithContext(ctx).Raw(`WITH candidate AS (SELECT r.id FROM receipts r JOIN events e ON e.id=r.event_id
 WHERE r.deleted_at IS NULL AND e.deleted_at IS NULL AND r.attempts<3 AND ((r.status='queued' AND r.available_at<=clock_timestamp()) OR (r.status='processing' AND r.lease_until<clock_timestamp()))
 ORDER BY r.sequence FOR UPDATE OF r SKIP LOCKED LIMIT 1)
 UPDATE receipts r SET status='processing',attempts=attempts+1,lease_token=?,lease_until=clock_timestamp()+interval '120 seconds',updated_at=clock_timestamp()
 FROM candidate c WHERE r.id=c.id RETURNING r.*`, token).Scan(&value).Error
	if err != nil {
		return nil, err
	}
	if value.ID == "" {
		return nil, nil
	}
	return decode(value)
}

// Complete отбрасывает ответ потерявшего аренду воркера и повторяет только ограниченное число попыток.
func (r *Repository) Complete(ctx context.Context, v *models.Receipt, result *models.ReceiptResult, processingErr error) error {
	fields := map[string]interface{}{"updated_at": time.Now(), "lease_token": nil, "lease_until": nil}
	if processingErr != nil {
		fields["status"] = "queued"
		fields["available_at"] = time.Now().Add(time.Duration(v.Attempts*v.Attempts) * 5 * time.Second)
		if v.Attempts >= 3 {
			fields["status"] = "failed"
		}
		fields["error"] = "Не удалось распознать чек. Проверьте фотографию; позиции можно добавить вручную."
		if errors.Is(processingErr, commonerrors.ErrOCRUnavailable) && time.Since(v.QueuedAt) < 15*time.Minute {
			fields["status"] = "queued"
			fields["error"] = ""
			fields["attempts"] = gorm.Expr("attempts-1")
			fields["available_at"] = time.Now().Add(3 * time.Second)
		}
	} else {
		data, err := json.Marshal(result)
		if err != nil {
			return err
		}
		fields["result"] = string(data)
		fields["status"] = "ready"
		fields["error"] = ""
		fields["revision"] = gorm.Expr("revision+1")
	}
	return r.db.WithContext(ctx).Table("receipts").Where("id=? AND status='processing' AND lease_token=? AND deleted_at IS NULL", v.ID, v.LeaseToken).Updates(fields).Error
}

// Garbage помечает истёкшие аренды и удаляет старые неподтверждённые черновики через outbox.
func (r *Repository) Garbage(ctx context.Context) ([]string, error) {
	if err := r.db.WithContext(ctx).Exec("WITH expired AS (UPDATE receipts SET image_available=false,status='failed',error='Загрузка фотографии прервана. Позиции можно добавить вручную.',revision=revision+1 WHERE deleted_at IS NULL AND status='uploading' AND lease_until<clock_timestamp() RETURNING image_key) INSERT INTO receipt_image_gc(image_key) SELECT image_key FROM expired ON CONFLICT DO NOTHING").Error; err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Exec("UPDATE receipts SET status='failed',lease_token=NULL,lease_until=NULL,error='Распознавание прервано. Добавьте позиции вручную.',revision=revision+1 WHERE deleted_at IS NULL AND status='processing' AND lease_until<clock_timestamp() AND attempts>=3").Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Exec(`WITH candidates AS (
 SELECT id FROM receipts WHERE status NOT IN ('confirmed','processing')
 AND updated_at<clock_timestamp()-interval '30 days' ORDER BY updated_at
 LIMIT 100 FOR UPDATE SKIP LOCKED
) DELETE FROM receipts r USING candidates c WHERE r.id=c.id
 AND r.status NOT IN ('confirmed','processing')
 AND r.updated_at<clock_timestamp()-interval '30 days'`).Error; err != nil {
		return nil, err
	}
	if err := r.db.WithContext(ctx).Exec("DELETE FROM receipt_admissions WHERE id IN (SELECT id FROM receipt_admissions WHERE created_at<clock_timestamp()-interval '7 days' ORDER BY created_at LIMIT 1000)").Error; err != nil {
		return nil, err
	}
	var keys []string
	err := r.db.WithContext(ctx).Table("receipt_image_gc").Where("created_at<clock_timestamp()-interval '1 hour'").Order("created_at").Limit(100).Pluck("image_key", &keys).Error
	return keys, err
}

// DeleteGarbage подтверждает успешное удаление объекта, допускающее повтор.
func (r *Repository) DeleteGarbage(ctx context.Context, key string) error {
	return dbctx.WithTx(ctx, r.db, func(ctx context.Context) error {
		tx := dbctx.GetTx(ctx, r.db).WithContext(ctx)
		if err := tx.Exec("UPDATE receipts SET image_available=false WHERE image_key=?", key).Error; err != nil {
			return err
		}
		return tx.Exec("DELETE FROM receipt_image_gc WHERE image_key=?", key).Error
	})
}

// FinishUpload публикует задачу только после сохранения фото и при действующей аренде загрузки.
func (r *Repository) FinishUpload(ctx context.Context, v *models.Receipt, success bool) error {
	status, message := "failed", "Загрузка фотографии прервана. Позиции можно добавить вручную."
	if success {
		status, message = "queued", ""
	}
	err := dbctx.WithTx(ctx, r.db, func(ctx context.Context) error {
		tx := dbctx.GetTx(ctx, r.db).WithContext(ctx)
		result := tx.Exec("UPDATE receipts SET status=?,error=?,image_available=?,lease_token=NULL,lease_until=NULL,revision=revision+1,updated_at=clock_timestamp() WHERE id=? AND status='uploading' AND lease_token=? AND lease_until>clock_timestamp() AND deleted_at IS NULL", status, message, success, v.ID, v.LeaseToken)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		if !success {
			return tx.Exec("INSERT INTO receipt_image_gc(image_key) VALUES(?) ON CONFLICT DO NOTHING", v.ImageKey).Error
		}
		return nil
	})
	if err != nil {
		return err
	}
	v.Status = status
	v.ImageAvailable = success
	v.Error = message
	v.Revision++
	return nil
}

// admit ограничивает суммарный бюджет загрузок и ручных повторов до записи задания.
func (r *Repository) admit(ctx context.Context, actorID int64) error {
	tx := dbctx.GetTx(ctx, r.db).WithContext(ctx)
	if err := tx.Exec("SELECT pg_advisory_xact_lock(20581407,2)").Error; err != nil {
		return err
	}
	var quota struct {
		Daily   int
		Pending int
		Global  int
	}
	if err := tx.Raw(`SELECT
 (SELECT count(*) FROM (SELECT 1 FROM receipt_admissions WHERE actor_id=? AND created_at>statement_timestamp()-interval '1 day' LIMIT 50) d) AS daily,
 (SELECT count(*) FROM (SELECT 1 FROM receipts r JOIN events e ON e.id=r.event_id WHERE r.queued_by=? AND r.deleted_at IS NULL AND e.deleted_at IS NULL AND r.status IN ('uploading','queued','processing') LIMIT 3) p) AS pending,
 (SELECT count(*) FROM (SELECT 1 FROM receipts r JOIN events e ON e.id=r.event_id WHERE r.deleted_at IS NULL AND e.deleted_at IS NULL AND r.status IN ('uploading','queued','processing') LIMIT 1000) g) AS global`, actorID, actorID).Scan(&quota).Error; err != nil {
		return err
	}
	if quota.Daily >= 50 || quota.Pending >= 3 || quota.Global >= 1000 {
		return commonerrors.NewValidationError("receipt", "очередь заполнена или достигнут лимит: три одновременных и 50 загрузок или повторов в сутки")
	}
	return nil
}

// Retry атомарно повторяет готовый черновик, сохраняя правки и журнал суточной квоты.
func (r *Repository) Retry(ctx context.Context, v *models.Receipt, actorID int64) error {
	tx := dbctx.GetTx(ctx, r.db).WithContext(ctx)
	var expired int64
	if err := tx.Table("receipt_image_gc").Where("image_key=?", v.ImageKey).Count(&expired).Error; err != nil {
		return err
	}
	if expired > 0 || !v.ImageAvailable {
		return commonerrors.NewValidationError("receipt", "фотография больше недоступна; загрузите её заново")
	}
	if err := r.admit(ctx, actorID); err != nil {
		return err
	}
	result := tx.Exec("UPDATE receipts SET status='queued',attempts=0,lease_token=NULL,lease_until=NULL,error='',queued_by=?,queued_at=clock_timestamp(),available_at=clock_timestamp(),retry_revision=?,revision=revision+1,updated_at=clock_timestamp() WHERE id=? AND revision=? AND deleted_at IS NULL AND status IN ('ready','failed')", actorID, v.Revision, v.ID, v.Revision)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return commonerrors.NewValidationError("receipt", "чек уже изменён; обновите его")
	}
	if err := tx.Exec("INSERT INTO receipt_admissions(receipt_id,actor_id) VALUES(?,?)", v.ID, actorID).Error; err != nil {
		return err
	}
	v.RetryRevision = v.Revision
	v.Revision++
	v.Status = "queued"
	v.Error = ""
	v.Attempts = 0
	v.LeaseToken = ""
	return nil
}

// Remove скрывает неподтверждённый чек и ставит фотографию в устойчивую очередь удаления.
func (r *Repository) Remove(ctx context.Context, v *models.Receipt) error {
	tx := dbctx.GetTx(ctx, r.db).WithContext(ctx)
	result := tx.Exec("UPDATE receipts SET image_available=false,deleted_at=clock_timestamp(),updated_at=clock_timestamp(),revision=revision+1,status='failed',lease_token=NULL,lease_until=NULL WHERE id=? AND revision=? AND status!='confirmed' AND deleted_at IS NULL", v.ID, v.Revision)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return commonerrors.NewValidationError("receipt", "чек уже изменён; обновите его")
	}
	return tx.Exec("INSERT INTO receipt_image_gc(image_key) VALUES(?) ON CONFLICT DO NOTHING", v.ImageKey).Error
}

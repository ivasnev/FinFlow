package history

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
	"time"
)

// Repository читает журнал, доступ к которому ограничен участием в мероприятии.
type Repository struct{ db *gorm.DB }

// NewRepository создаёт читатель журнала.
func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

// List возвращает ограниченную страницу с устойчивым курсором.
func (r *Repository) List(ctx context.Context, eventID, before int64) ([]models.HistoryEntry, error) {
	rows := []models.HistoryEntry{}
	q := r.db.WithContext(ctx).Where("event_id=?", eventID)
	if before > 0 {
		q = q.Where("id < ?", before)
	}
	err := q.Order("id DESC").Limit(50).Find(&rows).Error
	return rows, err
}

// Record вызывается в транзакции изменения; ошибка записи откатывает само изменение.
// Служебные вызовы без доверенного пользовательского контекста не подменяют автора.
func Record(ctx context.Context, tx *gorm.DB, eventID int64, kind string, entityID int64, before, after map[string]string) error {
	id, ok := actor.ID(ctx)
	if !ok {
		return nil
	}
	if before == nil {
		before = map[string]string{}
	}
	if after == nil {
		after = map[string]string{}
	}
	if reflect.DeepEqual(before, after) {
		return nil
	}
	var name string
	if err := tx.Raw("SELECT COALESCE(NULLIF(name_cashed,''),nickname_cashed,'Участник') FROM users WHERE user_id=? AND NOT is_dummy", id).Scan(&name).Error; err != nil {
		return err
	}
	if name == "" {
		name = "Участник"
	}
	b, err := json.Marshal(before)
	if err != nil {
		return err
	}
	a, err := json.Marshal(after)
	if err != nil {
		return err
	}
	return tx.Exec("INSERT INTO event_history(event_id,actor_id,actor_name,kind,entity_id,before_values,after_values) VALUES(?,?,?,?,?,?::jsonb,?::jsonb)", eventID, id, name, kind, entityID, string(b), string(a)).Error
}

// Expense фиксирует описание, деньги и исходные доли до их возможного удаления.
func Expense(tx *gorm.DB, id int) (map[string]string, error) {
	if _, ok := actor.ID(tx.Statement.Context); !ok {
		return nil, nil
	}
	var row struct {
		Name       string
		TotalPaid  float64
		Payer      string
		PayerID    int64
		CategoryID int
		Category   string
		SplitType  int
	}
	err := tx.Raw("SELECT t.name,t.total_paid,COALESCE(NULLIF(u.name_cashed,''),u.nickname_cashed,'Участник') AS payer,t.payer_id,COALESCE(t.transaction_category_id,0) AS category_id,COALESCE(c.name,'Без категории') AS category,t.split_type FROM transactions t LEFT JOIN users u ON u.id=t.payer_id LEFT JOIN transaction_categories c ON c.id=t.transaction_category_id WHERE t.id=?", id).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	var shares []struct {
		Name   string
		UserID int64
		Value  float64
	}
	err = tx.Raw("SELECT COALESCE(NULLIF(u.name_cashed,''),u.nickname_cashed,'Участник') AS name,s.user_id,s.value FROM transaction_shares s LEFT JOIN users u ON u.id=s.user_id WHERE s.transaction_id=? ORDER BY s.user_id", id).Scan(&shares).Error
	if err != nil {
		return nil, err
	}
	// ID типов совпадают с существующим калькулятором.
	types := map[int]string{0: "Поровну", 1: "По процентам", 2: "По суммам", 3: "По частям"}
	parts := make([]string, 0, len(shares))
	canonical := make([]string, 0, len(shares))
	for _, s := range shares {
		parts = append(parts, fmt.Sprintf("%s: %.2f ₽", s.Name, s.Value))
		canonical = append(canonical, fmt.Sprintf("%d:%d", s.UserID, int64(math.Round(s.Value*100))))
	}
	return map[string]string{"_payer_id": fmt.Sprint(row.PayerID), "_category_id": fmt.Sprint(row.CategoryID), "_shares": strings.Join(canonical, ","), "_split_type": fmt.Sprint(row.SplitType), "_amount_cents": fmt.Sprint(int64(math.Round(row.TotalPaid * 100))), "Категория": row.Category, "Название": row.Name, "Сумма": fmt.Sprintf("%.2f ₽", row.TotalPaid), "Плательщик": row.Payer, "Деление": types[row.SplitType], "Доли": strings.Join(parts, "; ")}, nil
}

// Event оставляет в истории только редактируемые поля события.
func Event(tx *gorm.DB, id int64) (map[string]string, error) {
	if _, ok := actor.ID(tx.Statement.Context); !ok {
		return nil, nil
	}
	var row struct {
		Name        string
		Description string
		Category    string
		CategoryID  int
	}
	err := tx.Raw("SELECT e.name,e.description,COALESCE(e.category_id,0) AS category_id,COALESCE(c.name,'Без категории') AS category FROM events e LEFT JOIN event_categories c ON c.id=e.category_id WHERE e.id=?", id).Scan(&row).Error
	return map[string]string{"_category_id": fmt.Sprint(row.CategoryID), "Название": row.Name, "Описание": row.Description, "Категория": row.Category}, err
}

// Payment сохраняет участников и сумму возврата независимо от последующих изменений профиля.
func Payment(tx *gorm.DB, row *models.Settlement) (map[string]string, error) {
	if _, ok := actor.ID(tx.Statement.Context); !ok {
		return nil, nil
	}
	var names []struct {
		ID   int64
		Name string
	}
	if err := tx.Raw("SELECT id,COALESCE(NULLIF(name_cashed,''),nickname_cashed,'Участник') AS name FROM users WHERE id IN (?,?)", row.FromUserID, row.ToUserID).Scan(&names).Error; err != nil {
		return nil, err
	}
	m := map[int64]string{}
	for _, n := range names {
		m[n.ID] = n.Name
	}
	return map[string]string{"_from_user_id": fmt.Sprint(row.FromUserID), "_to_user_id": fmt.Sprint(row.ToUserID), "_amount_cents": fmt.Sprint(row.AmountCents), "Отправитель": m[row.FromUserID], "Получатель": m[row.ToUserID], "Сумма": fmt.Sprintf("%.2f ₽", float64(row.AmountCents)/100)}, nil
}

// LockEvent задаёт единый порядок блокировок: мероприятие, затем его изменяемая сущность.
func LockEvent(tx *gorm.DB, eventID int64) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	var row struct {
		ID        int64
		DeletedAt *time.Time
	}
	result := tx.Raw("SELECT id,deleted_at FROM events WHERE id=? FOR UPDATE", eventID).Scan(&row)
	if result.Error != nil {
		return result.Error
	}
	if row.ID == 0 {
		return gorm.ErrRecordNotFound
	}
	if row.DeletedAt != nil {
		return errors.ErrEventArchived
	}
	return nil
}

// LockExpense удерживает снимок расхода и долей до записи изменения и его истории.
func LockExpense(tx *gorm.DB, id int) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	var eventID int64
	if err := tx.Raw("SELECT event_id FROM transactions WHERE id=?", id).Scan(&eventID).Error; err != nil {
		return err
	}
	if err := LockEvent(tx, eventID); err != nil {
		return err
	}
	return tx.Exec("SELECT id FROM transactions WHERE id=? FOR UPDATE", id).Error
}

// Task сохраняет значения задачи и имя исполнителя на момент изменения.
func Task(tx *gorm.DB, id uint) (map[string]string, error) {
	if _, ok := actor.ID(tx.Statement.Context); !ok {
		return nil, nil
	}
	var row struct {
		Title       string
		Description string
		Priority    int
		Assignee    string
		AssigneeID  int64
	}
	err := tx.Raw("SELECT t.title,t.description,t.priority,COALESCE(t.user_id,0) AS assignee_id,COALESCE(NULLIF(u.name_cashed,''),u.nickname_cashed,'Участник') AS assignee FROM tasks t LEFT JOIN users u ON u.id=t.user_id WHERE t.id=?", id).Scan(&row).Error
	return map[string]string{"_assignee_id": fmt.Sprint(row.AssigneeID), "Название": row.Title, "Описание": row.Description, "Исполнитель": row.Assignee, "Приоритет": fmt.Sprint(row.Priority)}, err
}

// Member добавляет запись только о фактически изменённой связи участия.
func Member(ctx context.Context, tx *gorm.DB, eventID, userID int64, added bool) error {
	if _, ok := actor.ID(ctx); !ok {
		return nil
	}
	var name string
	if err := tx.Raw("SELECT COALESCE(NULLIF(name_cashed,''),nickname_cashed,'Участник') FROM users WHERE id=?", userID).Scan(&name).Error; err != nil {
		return err
	}
	values := map[string]string{"_member_id": fmt.Sprint(userID), "Участник": name}
	if added {
		return Record(ctx, tx, eventID, "member.added", userID, nil, values)
	}
	return Record(ctx, tx, eventID, "member.removed", userID, values, nil)
}

package event

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/ivasnev/FinFlow/ff-split/internal/common/db"
	customErrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Используем модели БД и мапперы для преобразования

// EventRepository реализует интерфейс repository.Event
type EventRepository struct {
	db *gorm.DB
}

// NewEventRepository создает новый экземпляр EventRepository
func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{
		db: db,
	}
}

// GetAll возвращает все мероприятия
func (r *EventRepository) GetAll(ctx context.Context) ([]models.Event, error) {
	var dbEvents []Event
	err := r.db.WithContext(ctx).Select("events.*,event_purge_at(events.deleted_at) AS purge_at").Where("deleted_at IS NULL").Find(&dbEvents).Error
	if err != nil {
		return nil, err
	}
	return extractSlice(dbEvents), nil
}

// GetByID возвращает мероприятие по ID
func (r *EventRepository) GetByID(ctx context.Context, id int64) (*models.Event, error) {
	var dbEvent Event
	err := db.GetTx(ctx, r.db).WithContext(ctx).Select("events.*,event_purge_at(deleted_at) AS purge_at").First(&dbEvent, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // возвращаем nil, nil если мероприятие не найдено
		}
		return nil, err
	}
	return extract(&dbEvent), nil
}

// Create создает новое мероприятие
func (r *EventRepository) Create(ctx context.Context, event *models.Event) error {
	dbEvent := load(event)
	if err := db.GetTx(ctx, r.db).WithContext(ctx).Create(dbEvent).Error; err != nil {
		return err
	}
	// Обновляем ID в оригинальной модели
	event.ID = dbEvent.ID
	return nil
}

// Update сохраняет выбранные поля, включая нулевые значения, и возвращает фактически записанную строку в event.
// Без списка колонок заменяет редактируемые имя, описание и категорию; фото и статус сохраняются.
func (r *EventRepository) Update(ctx context.Context, id int64, event *models.Event, columns ...string) error {
	if len(columns) == 0 {
		columns = []string{"name", "description", "category_id"}
	}
	for _, column := range columns {
		if column != "name" && column != "description" && column != "category_id" {
			return fmt.Errorf("недопустимое поле мероприятия: %s", column)
		}
	}
	dbEvent := load(event)
	dbEvent.ID = id
	result := db.GetTx(ctx, r.db).WithContext(ctx).Model(dbEvent).Where("id = ?", id).
		Clauses(clause.Returning{}).Select(columns).Updates(dbEvent)
	if result.Error != nil {
		return fmt.Errorf("ошибка при обновлении мероприятия: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	*event = *extract(dbEvent)
	return nil
}

// GetByUserID возвращает мероприятия пользователя
func (r *EventRepository) GetByUserID(ctx context.Context, userID int64) ([]models.Event, error) {
	var dbEvents []Event
	err := r.db.WithContext(ctx).Select("events.*,event_purge_at(events.deleted_at) AS purge_at").
		Joins("JOIN user_event ON events.id = user_event.event_id").
		Where("user_event.user_id = ? AND events.deleted_at IS NULL", userID).
		Find(&dbEvents).Error
	if err != nil {
		return nil, err
	}
	return extractSlice(dbEvents), nil
}

// CalculateUserBalances рассчитывает баланс пользователя по событиям
func (r *EventRepository) CalculateUserBalances(ctx context.Context, userID int64, eventIDs []int64) (map[int64]float64, error) {
	if len(eventIDs) == 0 {
		return make(map[int64]float64), nil
	}

	type BalanceResult struct {
		EventID int64   `gorm:"column:event_id"`
		Balance float64 `gorm:"column:balance"`
	}

	var results []BalanceResult

	// Используем прямой SQL запрос для расчета баланса
	// Баланс = что ему должны (to_user_id = userID) - что он должен (from_user_id = userID)
	err := r.db.WithContext(ctx).
		Raw(`SELECT event_id, COALESCE(SUM(CASE WHEN to_user_id=? THEN amount ELSE 0 END),0)
        - COALESCE(SUM(CASE WHEN from_user_id=? THEN amount ELSE 0 END),0) AS balance
        FROM (
          SELECT t.event_id,d.from_user_id,d.to_user_id,d.amount FROM debts d JOIN transactions t ON d.transaction_id=t.id WHERE t.event_id IN ?
          UNION ALL SELECT event_id,to_user_id,from_user_id,amount_cents::numeric/100 FROM settlements WHERE event_id IN ? AND canceled_at IS NULL
        ) ledger GROUP BY event_id`, userID, userID, eventIDs, eventIDs).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	balances := make(map[int64]float64, len(results))
	for _, result := range results {
		balances[result.EventID] = result.Balance
	}

	// Инициализируем нулевой баланс для событий без долгов
	for _, eventID := range eventIDs {
		if _, exists := balances[eventID]; !exists {
			balances[eventID] = 0
		}
	}

	return balances, nil
}

// Delete удаляет мероприятие и все связанные данные
func (r *EventRepository) Delete(ctx context.Context, id int64) error {
	err := db.WithTx(ctx, r.db, func(ctx context.Context) error {
		tx := db.GetTx(ctx, r.db).WithContext(ctx)

		// Сначала удаляем долги (они ссылаются на транзакции)
		result := tx.Exec("DELETE FROM debts WHERE transaction_id IN (SELECT id FROM transactions WHERE event_id = ?)", id)
		if result.Error != nil {
			return result.Error
		}

		// Удаляем доли транзакций
		result = tx.Exec("DELETE FROM transaction_shares WHERE transaction_id IN (SELECT id FROM transactions WHERE event_id = ?)", id)
		if result.Error != nil {
			return result.Error
		}

		// Удаляем транзакции
		result = tx.Exec("DELETE FROM transactions WHERE event_id = ?", id)
		if result.Error != nil {
			return result.Error
		}

		// Удаляем оптимизированные долги
		result = tx.Exec("DELETE FROM optimized_debts WHERE event_id = ?", id)
		if result.Error != nil {
			return result.Error
		}

		// Удаляем активности
		result = tx.Exec("DELETE FROM activities WHERE event_id = ?", id)
		if result.Error != nil {
			return result.Error
		}

		// Удаляем задачи
		result = tx.Exec("DELETE FROM tasks WHERE event_id = ?", id)
		if result.Error != nil {
			return result.Error
		}

		// Удаляем связи пользователей с мероприятием
		result = tx.Exec("DELETE FROM user_event WHERE event_id = ?", id)
		if result.Error != nil {
			return result.Error
		}

		// Наконец, удаляем само мероприятие
		result = tx.Delete(&Event{}, id)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return customErrors.NewEntityNotFoundError(strconv.FormatInt(id, 10), "event")
		}

		return nil
	})
	return err
}

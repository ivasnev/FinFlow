package transaction

import (
	"context"
	"errors"

	"gorm.io/gorm"

	domainerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
)

// TransactionRepository репозиторий для работы с транзакциями
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository создает новый репозиторий транзакций
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// WithTx привязывает новый экземпляр к транзакции, не меняя общий репозиторий.
func (r *TransactionRepository) WithTx(tx *gorm.DB) repository.Transaction {
	return NewTransactionRepository(tx)
}

// WithContext создаёт экземпляр с отменяемыми запросами без изменения общего репозитория.
func (r *TransactionRepository) WithContext(ctx context.Context) repository.Transaction {
	return NewTransactionRepository(r.db.WithContext(ctx))
}

// InvalidateOptimization дёшево помечает результат устаревшим в транзакции изменения расхода.
func (r *TransactionRepository) InvalidateOptimization(eventID int64) error {
	return r.db.Table("events").Where("id = ?", eventID).UpdateColumn("debt_revision", gorm.Expr("debt_revision + 1")).Error
}

// GetOptimizationState учитывает также корректный пустой результат оптимизации.
func (r *TransactionRepository) GetOptimizationState(eventID int64) (*models.DebtOptimizationState, error) {
	var state models.DebtOptimizationState
	if err := r.db.Table("events").Select("debt_revision, optimized_debt_revision").Where("id = ?", eventID).Take(&state).Error; err != nil {
		return nil, err
	}
	return &state, nil
}

// RunOptimization удерживает замок только между вычислениями одного мероприятия, не блокируя запись расходов.
func (r *TransactionRepository) RunOptimization(ctx context.Context, eventID int64, fn func(repository.Transaction) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked bool
		if err := tx.Raw("SELECT pg_try_advisory_xact_lock(?)", eventID).Scan(&locked).Error; err != nil {
			return err
		}
		if !locked {
			return domainerrors.ErrOptimizationBusy
		}
		return fn(r.WithTx(tx))
	})
}

// GetTransactionsByEventID возвращает список транзакций мероприятия
func (r *TransactionRepository) GetTransactionsByEventID(eventID int64) ([]models.Transaction, error) {
	var dbTransactions []Transaction
	if err := r.db.Where("event_id = ?", eventID).Find(&dbTransactions).Error; err != nil {
		return nil, err
	}
	return extractSlice(dbTransactions), nil
}

// GetTransactionByID возвращает транзакцию по ID
func (r *TransactionRepository) GetTransactionByID(id int) (*models.Transaction, error) {
	var dbTransaction Transaction
	if err := r.db.First(&dbTransaction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("транзакция не найдена")
		}
		return nil, err
	}
	return extract(&dbTransaction), nil
}

// CreateTransaction создает новую транзакцию
func (r *TransactionRepository) CreateTransaction(tx *models.Transaction) error {
	dbTx := load(tx)
	if err := r.db.Create(dbTx).Error; err != nil {
		return err
	}
	tx.ID = dbTx.ID
	return nil
}

// UpdateTransaction обновляет существующую транзакцию
func (r *TransactionRepository) UpdateTransaction(tx *models.Transaction) error {
	dbTx := load(tx)
	result := r.db.Save(dbTx)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("транзакция не найдена")
	}
	return nil
}

// DeleteTransaction удаляет транзакцию и связанные с ней доли и долги
func (r *TransactionRepository) DeleteTransaction(id int) error {
	// Выполняем операции в транзакции
	return r.db.Transaction(func(tx *gorm.DB) error {
		var expense Transaction
		if err := tx.First(&expense, id).Error; err != nil {
			return err
		}
		// Удаляем связанные долги
		if err := tx.Where("transaction_id = ?", id).Delete(&Debt{}).Error; err != nil {
			return err
		}

		// Удаляем связанные доли
		if err := tx.Where("transaction_id = ?", id).Delete(&TransactionShare{}).Error; err != nil {
			return err
		}

		// Удаляем саму транзакцию
		result := tx.Delete(&Transaction{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("транзакция не найдена")
		}
		if expense.EventID != nil {
			if err := r.WithTx(tx).InvalidateOptimization(*expense.EventID); err != nil {
				return err
			}
		}

		return nil
	})
}

// GetSharesByTransactionID возвращает доли пользователей в транзакции
func (r *TransactionRepository) GetSharesByTransactionID(transactionID int) ([]models.TransactionShare, error) {
	var dbShares []TransactionShare
	if err := r.db.Where("transaction_id = ?", transactionID).Find(&dbShares).Error; err != nil {
		return nil, err
	}
	// Преобразуем в бизнес-модели
	shares := make([]models.TransactionShare, len(dbShares))
	for i, dbShare := range dbShares {
		if extracted := extractTransactionShare(&dbShare); extracted != nil {
			shares[i] = *extracted
		}
	}
	return shares, nil
}

// GetDebtsByTransactionID возвращает долги в рамках транзакции
func (r *TransactionRepository) GetDebtsByTransactionID(transactionID int) ([]models.Debt, error) {
	var dbDebts []Debt
	if err := r.db.Where("transaction_id = ?", transactionID).Find(&dbDebts).Error; err != nil {
		return nil, err
	}
	return extractDebtSlice(dbDebts), nil
}

// GetDebtsByEventID возвращает долги в рамках мероприятия
func (r *TransactionRepository) GetDebtsByEventID(eventID int64) ([]models.Debt, error) {
	var rows []struct {
		ID            int
		TransactionID int
		FromUserID    int64
		ToUserID      int64
		Amount        float64
		Settlement    bool
	}
	// Один SQL-снимок объединяет расходы и возвраты, не умножая их JOIN-ом.
	err := r.db.Table(`(SELECT d.id,d.transaction_id,d.from_user_id,d.to_user_id,d.amount,false AS settlement
 FROM debts d JOIN transactions t ON t.id=d.transaction_id WHERE t.event_id=?
 UNION ALL SELECT 0,0,to_user_id,from_user_id,amount_cents::numeric/100,true
 FROM settlements WHERE event_id=? AND canceled_at IS NULL) AS debts`, eventID, eventID).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]models.Debt, 0, len(rows))
	for _, row := range rows {
		result = append(result, models.Debt{ID: row.ID, TransactionID: row.TransactionID, FromUserID: row.FromUserID, ToUserID: row.ToUserID, Amount: row.Amount, Settlement: row.Settlement})
	}
	return result, nil
}

// GetDebtsByEventIDToUser учитывает остатки после возвратов, включая оплату непрямому кредитору.
func (r *TransactionRepository) GetDebtsByEventIDToUser(eventID, userID int64) ([]models.Debt, error) {
	rows, err := r.GetDebtsByEventID(eventID)
	if err != nil {
		return nil, err
	}
	result := []models.Debt{}
	for _, row := range models.OutstandingDebts(rows) {
		if row.ToUserID == userID {
			result = append(result, row)
		}
	}
	return result, nil
}

// GetDebtsByEventIDFromUser возвращает текущие исходящие долги.
func (r *TransactionRepository) GetDebtsByEventIDFromUser(eventID, userID int64) ([]models.Debt, error) {
	rows, err := r.GetDebtsByEventID(eventID)
	if err != nil {
		return nil, err
	}
	result := []models.Debt{}
	for _, row := range models.OutstandingDebts(rows) {
		if row.FromUserID == userID {
			result = append(result, row)
		}
	}
	return result, nil
}

// CreateTransactionShares создает доли пользователей в транзакции
func (r *TransactionRepository) CreateTransactionShares(shares []models.TransactionShare) error {
	if len(shares) == 0 {
		return nil // Нет долей для создания
	}
	dbShares := make([]TransactionShare, len(shares))
	for i, share := range shares {
		dbShares[i] = *loadTransactionShare(&share)
	}
	return r.db.Create(&dbShares).Error
}

// CreateDebts создает долги между пользователями
func (r *TransactionRepository) CreateDebts(debts []models.Debt) error {
	if len(debts) == 0 {
		return nil // Нет долгов для создания
	}
	dbDebts := make([]Debt, len(debts))
	for i, debt := range debts {
		dbDebts[i] = *loadDebt(&debt)
	}
	return r.db.Create(&dbDebts).Error
}

// DeleteSharesByTransactionID удаляет все доли в транзакции
func (r *TransactionRepository) DeleteSharesByTransactionID(transactionID int) error {
	return r.db.Where("transaction_id = ?", transactionID).Delete(&TransactionShare{}).Error
}

// DeleteDebtsByTransactionID удаляет все долги в транзакции
func (r *TransactionRepository) DeleteDebtsByTransactionID(transactionID int) error {
	return r.db.Where("transaction_id = ?", transactionID).Delete(&Debt{}).Error
}

// GetOptimizedDebtsByEventID возвращает оптимизированные долги по ID мероприятия
func (r *TransactionRepository) GetOptimizedDebtsByEventID(eventID int64) ([]models.OptimizedDebt, error) {
	var dbDebts []OptimizedDebt
	if err := r.db.Where("event_id = ?", eventID).Find(&dbDebts).Error; err != nil {
		return nil, err
	}
	return extractOptimizedDebtSlice(dbDebts), nil
}

// GetOptimizedDebtsByEventIDWithUsers возвращает оптимизированные долги по ID мероприятия с загрузкой пользователей
func (r *TransactionRepository) GetOptimizedDebtsByEventIDWithUsers(eventID int64) ([]models.OptimizedDebt, error) {
	var dbDebts []OptimizedDebt
	if err := r.db.Where("event_id = ?", eventID).Find(&dbDebts).Error; err != nil {
		return nil, err
	}
	return extractOptimizedDebtSlice(dbDebts), nil
}

// GetOptimizedDebtsByUserID возвращает оптимизированные долги по ID пользователя в мероприятии
func (r *TransactionRepository) GetOptimizedDebtsByUserID(eventID, userID int64) ([]models.OptimizedDebt, error) {
	var dbDebts []OptimizedDebt
	if err := r.db.Where("event_id = ? AND (from_user_id = ? OR to_user_id = ?)", eventID, userID, userID).Find(&dbDebts).Error; err != nil {
		return nil, err
	}
	return extractOptimizedDebtSlice(dbDebts), nil
}

// GetOptimizedDebtsByUserIDWithUsers возвращает оптимизированные долги по ID пользователя в мероприятии с загрузкой пользователей
func (r *TransactionRepository) GetOptimizedDebtsByUserIDWithUsers(eventID, userID int64) ([]models.OptimizedDebt, error) {
	var dbDebts []OptimizedDebt
	if err := r.db.Where("event_id = ? AND (from_user_id = ? OR to_user_id = ?)", eventID, userID, userID).Find(&dbDebts).Error; err != nil {
		return nil, err
	}
	return extractOptimizedDebtSlice(dbDebts), nil
}

// SaveOptimizedDebts сохраняет оптимизированные долги для мероприятия (удаляет старые и сохраняет новые)
func (r *TransactionRepository) SaveOptimizedDebts(eventID int64, revision int64, debts []models.OptimizedDebt) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Сравнение и публикация версии происходят атомарно с заменой записей кэша.
		updated := tx.Table("events").Where("id = ? AND debt_revision = ?", eventID, revision).UpdateColumn("optimized_debt_revision", revision)
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != 1 {
			return domainerrors.ErrDebtRevisionChanged
		}
		// Удаляем старые оптимизированные долги
		if err := tx.Where("event_id = ?", eventID).Delete(&OptimizedDebt{}).Error; err != nil {
			return err
		}

		// Сохраняем новые оптимизированные долги
		if len(debts) > 0 {
			dbDebts := make([]OptimizedDebt, len(debts))
			for i, debt := range debts {
				dbDebts[i] = *loadOptimizedDebt(&debt)
			}
			if err := tx.Create(&dbDebts).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// DeleteOptimizedDebtsByEventID удаляет оптимизированные долги по ID мероприятия
func (r *TransactionRepository) DeleteOptimizedDebtsByEventID(eventID int64) error {
	result := r.db.Where("event_id = ?", eventID).Delete(&OptimizedDebt{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("оптимизированные долги не найдены")
	}
	return nil
}

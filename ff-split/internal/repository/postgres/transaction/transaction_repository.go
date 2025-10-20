package transaction

import (
	"errors"

	"gorm.io/gorm"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// TransactionRepository репозиторий для работы с транзакциями
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository создает новый репозиторий транзакций
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
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
	var dbDebts []Debt
	if err := r.db.Joins("JOIN transactions ON transactions.id = debts.transaction_id").
		Where("transactions.event_id = ?", eventID).
		Find(&dbDebts).Error; err != nil {
		return nil, err
	}
	return extractDebtSlice(dbDebts), nil
}

// GetDebtsByEventIDToUser возвращает долги в рамках мероприятия для конкретного пользователя
func (r *TransactionRepository) GetDebtsByEventIDToUser(eventID int64, userID int64) ([]models.Debt, error) {
	var dbDebts []Debt
	if err := r.db.Joins("JOIN transactions ON transactions.id = debts.transaction_id").
		Where("transactions.event_id = ? AND debts.to_user_id = ?", eventID, userID).
		Find(&dbDebts).Error; err != nil {
		return nil, err
	}
	return extractDebtSlice(dbDebts), nil
}

// GetDebtsByEventIDFromUser возвращает долги в рамках мероприятия для конкретного пользователя
func (r *TransactionRepository) GetDebtsByEventIDFromUser(eventID int64, userID int64) ([]models.Debt, error) {
	var dbDebts []Debt
	if err := r.db.Joins("JOIN transactions ON transactions.id = debts.transaction_id").
		Where("transactions.event_id = ? AND debts.from_user_id = ?", eventID, userID).
		Find(&dbDebts).Error; err != nil {
		return nil, err
	}
	return extractDebtSlice(dbDebts), nil
}

// CreateTransactionShares создает доли пользователей в транзакции
func (r *TransactionRepository) CreateTransactionShares(shares []models.TransactionShare) error {
	dbShares := make([]TransactionShare, len(shares))
	for i, share := range shares {
		dbShares[i] = *loadTransactionShare(&share)
	}
	return r.db.Create(&dbShares).Error
}

// CreateDebts создает долги между пользователями
func (r *TransactionRepository) CreateDebts(debts []models.Debt) error {
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
func (r *TransactionRepository) SaveOptimizedDebts(eventID int64, debts []models.OptimizedDebt) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
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

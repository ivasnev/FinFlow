package postgres

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
	var transactions []models.Transaction
	if err := r.db.Where("event_id = ?", eventID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTransactionByID возвращает транзакцию по ID
func (r *TransactionRepository) GetTransactionByID(id int) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := r.db.First(&transaction, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("транзакция не найдена")
		}
		return nil, err
	}
	return &transaction, nil
}

// CreateTransaction создает новую транзакцию
func (r *TransactionRepository) CreateTransaction(tx *models.Transaction) error {
	return r.db.Create(tx).Error
}

// UpdateTransaction обновляет существующую транзакцию
func (r *TransactionRepository) UpdateTransaction(tx *models.Transaction) error {
	result := r.db.Save(tx)
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
		if err := tx.Where("transaction_id = ?", id).Delete(&models.Debt{}).Error; err != nil {
			return err
		}

		// Удаляем связанные доли
		if err := tx.Where("transaction_id = ?", id).Delete(&models.TransactionShare{}).Error; err != nil {
			return err
		}

		// Удаляем саму транзакцию
		result := tx.Delete(&models.Transaction{}, id)
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
	var shares []models.TransactionShare
	if err := r.db.Where("transaction_id = ?", transactionID).Find(&shares).Error; err != nil {
		return nil, err
	}
	return shares, nil
}

// GetDebtsByTransactionID возвращает долги в рамках транзакции
func (r *TransactionRepository) GetDebtsByTransactionID(transactionID int) ([]models.Debt, error) {
	var debts []models.Debt
	if err := r.db.Where("transaction_id = ?", transactionID).Find(&debts).Error; err != nil {
		return nil, err
	}
	return debts, nil
}

// GetDebtsByEventID возвращает долги в рамках мероприятия
func (r *TransactionRepository) GetDebtsByEventID(eventID int64) ([]models.Debt, error) {
	var debts []models.Debt
	if err := r.db.Preload("FromUser").Preload("ToUser").Joins("JOIN transactions ON transactions.id = debts.transaction_id").
		Where("transactions.event_id = ?", eventID).
		Find(&debts).Error; err != nil {
		return nil, err
	}
	return debts, nil
}

// GetDebtsByEventIDToUser возвращает долги в рамках мероприятия для конкретного пользователя
func (r *TransactionRepository) GetDebtsByEventIDToUser(eventID int64, userID int64) ([]models.Debt, error) {
	var debts []models.Debt
	if err := r.db.Joins("JOIN transactions ON transactions.id = debts.transaction_id").
		Where("transactions.event_id = ? AND debts.to_user_id = ?", eventID, userID).
		Preload("FromUser").Find(&debts).Error; err != nil {
		return nil, err
	}
	return debts, nil
}

// GetDebtsByEventIDFromUser возвращает долги в рамках мероприятия для конкретного пользователя
func (r *TransactionRepository) GetDebtsByEventIDFromUser(eventID int64, userID int64) ([]models.Debt, error) {
	var debts []models.Debt
	if err := r.db.Joins("JOIN transactions ON transactions.id = debts.transaction_id").
		Where("transactions.event_id = ? AND debts.from_user_id = ?", eventID, userID).
		Preload("ToUser").Find(&debts).Error; err != nil {
		return nil, err
	}
	return debts, nil
}

// CreateTransactionShares создает доли пользователей в транзакции
func (r *TransactionRepository) CreateTransactionShares(shares []models.TransactionShare) error {
	return r.db.Create(&shares).Error
}

// CreateDebts создает долги между пользователями
func (r *TransactionRepository) CreateDebts(debts []models.Debt) error {
	return r.db.Create(&debts).Error
}

// DeleteSharesByTransactionID удаляет все доли в транзакции
func (r *TransactionRepository) DeleteSharesByTransactionID(transactionID int) error {
	return r.db.Where("transaction_id = ?", transactionID).Delete(&models.TransactionShare{}).Error
}

// DeleteDebtsByTransactionID удаляет все долги в транзакции
func (r *TransactionRepository) DeleteDebtsByTransactionID(transactionID int) error {
	return r.db.Where("transaction_id = ?", transactionID).Delete(&models.Debt{}).Error
}

// GetOptimizedDebtsByEventID возвращает оптимизированные долги по ID мероприятия
func (r *TransactionRepository) GetOptimizedDebtsByEventID(eventID int64) ([]models.OptimizedDebt, error) {
	var debts []models.OptimizedDebt
	if err := r.db.Where("event_id = ?", eventID).Find(&debts).Error; err != nil {
		return nil, err
	}
	return debts, nil
}

// GetOptimizedDebtsByUserID возвращает оптимизированные долги по ID пользователя в мероприятии
func (r *TransactionRepository) GetOptimizedDebtsByUserID(eventID, userID int64) ([]models.OptimizedDebt, error) {
	var debts []models.OptimizedDebt
	if err := r.db.Where("event_id = ? AND (from_user_id = ? OR to_user_id = ?)", eventID, userID, userID).Find(&debts).Error; err != nil {
		return nil, err
	}
	return debts, nil
}

// SaveOptimizedDebts сохраняет оптимизированные долги для мероприятия (удаляет старые и сохраняет новые)
func (r *TransactionRepository) SaveOptimizedDebts(eventID int64, debts []models.OptimizedDebt) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Удаляем старые оптимизированные долги
		if err := tx.Where("event_id = ?", eventID).Delete(&models.OptimizedDebt{}).Error; err != nil {
			return err
		}

		// Сохраняем новые оптимизированные долги
		if len(debts) > 0 {
			if err := tx.Create(&debts).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// DeleteOptimizedDebtsByEventID удаляет оптимизированные долги по ID мероприятия
func (r *TransactionRepository) DeleteOptimizedDebtsByEventID(eventID int64) error {
	result := r.db.Where("event_id = ?", eventID).Delete(&models.OptimizedDebt{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("оптимизированные долги не найдены")
	}
	return nil
}

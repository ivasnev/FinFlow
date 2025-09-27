package postgres

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// TransactionRepository представляет собой репозиторий для работы с транзакциями в PostgreSQL
type TransactionRepository struct {
	db *gorm.DB
}

// NewTransactionRepository создает новый экземпляр TransactionRepository
func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

// GetTransactionsByEventID возвращает все транзакции по ID мероприятия
func (r *TransactionRepository) GetTransactionsByEventID(ctx context.Context, eventID int64) ([]models.Transaction, error) {
	var transactions []models.Transaction
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&transactions).Error; err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTransactionByID возвращает транзакцию по ID
func (r *TransactionRepository) GetTransactionByID(ctx context.Context, id int) (models.Transaction, error) {
	var transaction models.Transaction
	if err := r.db.WithContext(ctx).First(&transaction, id).Error; err != nil {
		return models.Transaction{}, err
	}
	return transaction, nil
}

// CreateTransaction создает новую транзакцию
func (r *TransactionRepository) CreateTransaction(ctx context.Context, transaction models.Transaction) (models.Transaction, error) {
	if err := r.db.WithContext(ctx).Create(&transaction).Error; err != nil {
		return models.Transaction{}, err
	}
	return transaction, nil
}

// UpdateTransaction обновляет существующую транзакцию
func (r *TransactionRepository) UpdateTransaction(ctx context.Context, transaction models.Transaction) error {
	return r.db.WithContext(ctx).Save(&transaction).Error
}

// DeleteTransaction удаляет транзакцию
func (r *TransactionRepository) DeleteTransaction(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Transaction{}, id).Error
}

// AddUserToTransaction добавляет пользователя в транзакцию
func (r *TransactionRepository) AddUserToTransaction(ctx context.Context, userTransaction models.UserTransaction) error {
	return r.db.WithContext(ctx).Create(&userTransaction).Error
}

// RemoveUserFromTransaction удаляет пользователя из транзакции
func (r *TransactionRepository) RemoveUserFromTransaction(ctx context.Context, userID int64, transactionID int) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND transaction_id = ?", userID, transactionID).
		Delete(&models.UserTransaction{}).Error
}

// GetTransactionUsers возвращает всех пользователей транзакции
func (r *TransactionRepository) GetTransactionUsers(ctx context.Context, transactionID int) ([]models.UserTransaction, error) {
	var userTransactions []models.UserTransaction
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).Find(&userTransactions).Error; err != nil {
		return nil, err
	}
	return userTransactions, nil
}

// GetTemporalTransactionsByEventID возвращает временные данные о транзакциях по ID мероприятия
func (r *TransactionRepository) GetTemporalTransactionsByEventID(ctx context.Context, eventID int64) ([]models.EventTransactionTemporalResponse, error) {
	var result []models.EventTransactionTemporalResponse

	// Здесь будет сложный запрос для вычисления кто кому сколько должен
	// Вместо этого в данном примере просто возвращаем пустой массив
	// В реальном приложении нужно будет реализовать логику расчета

	return result, nil
}

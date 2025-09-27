package postgres

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// TransactionTypeRepository представляет собой репозиторий для работы с типами транзакций в PostgreSQL
type TransactionTypeRepository struct {
	db *gorm.DB
}

// NewTransactionTypeRepository создает новый экземпляр TransactionTypeRepository
func NewTransactionTypeRepository(db *gorm.DB) *TransactionTypeRepository {
	return &TransactionTypeRepository{
		db: db,
	}
}

// GetTransactionTypes возвращает все типы транзакций
func (r *TransactionTypeRepository) GetTransactionTypes(ctx context.Context) ([]models.TransactionType, error) {
	var transactionTypes []models.TransactionType
	if err := r.db.WithContext(ctx).Find(&transactionTypes).Error; err != nil {
		return nil, err
	}
	return transactionTypes, nil
}

// GetTransactionTypeByID возвращает тип транзакции по ID
func (r *TransactionTypeRepository) GetTransactionTypeByID(ctx context.Context, id int) (models.TransactionType, error) {
	var transactionType models.TransactionType
	if err := r.db.WithContext(ctx).First(&transactionType, id).Error; err != nil {
		return models.TransactionType{}, err
	}
	return transactionType, nil
}

// CreateTransactionType создает новый тип транзакции
func (r *TransactionTypeRepository) CreateTransactionType(ctx context.Context, transactionType models.TransactionType) (models.TransactionType, error) {
	if err := r.db.WithContext(ctx).Create(&transactionType).Error; err != nil {
		return models.TransactionType{}, err
	}
	return transactionType, nil
}

// UpdateTransactionType обновляет существующий тип транзакции
func (r *TransactionTypeRepository) UpdateTransactionType(ctx context.Context, transactionType models.TransactionType) error {
	return r.db.WithContext(ctx).Save(&transactionType).Error
}

// DeleteTransactionType удаляет тип транзакции
func (r *TransactionTypeRepository) DeleteTransactionType(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.TransactionType{}, id).Error
}

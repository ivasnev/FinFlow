package repository

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Transaction определяет методы для работы с транзакциями
type Transaction interface {
	// Получение транзакций
	GetTransactionsByEventID(eventID int64) ([]models.Transaction, error)
	GetTransactionByID(id int) (*models.Transaction, error)

	// Управление транзакциями
	CreateTransaction(tx *models.Transaction) error
	UpdateTransaction(tx *models.Transaction) error
	DeleteTransaction(id int) error

	// Работа с долями транзакций
	GetSharesByTransactionID(transactionID int) ([]models.TransactionShare, error)
	CreateTransactionShares(shares []models.TransactionShare) error
	DeleteSharesByTransactionID(transactionID int) error

	// Работа с долгами
	GetDebtsByTransactionID(transactionID int) ([]models.Debt, error)
	GetDebtsByEventID(eventID int64) ([]models.Debt, error)
	GetDebtsByEventIDFromUser(eventID int64, userID int64) ([]models.Debt, error)
	GetDebtsByEventIDToUser(eventID int64, userID int64) ([]models.Debt, error)
	CreateDebts(debts []models.Debt) error
	DeleteDebtsByTransactionID(transactionID int) error

	// Работа с оптимизированными долгами
	GetOptimizedDebtsByEventID(eventID int64) ([]models.OptimizedDebt, error)
	GetOptimizedDebtsByEventIDWithUsers(eventID int64) ([]models.OptimizedDebt, error)
	GetOptimizedDebtsByUserID(eventID, userID int64) ([]models.OptimizedDebt, error)
	GetOptimizedDebtsByUserIDWithUsers(eventID, userID int64) ([]models.OptimizedDebt, error)
	SaveOptimizedDebts(eventID int64, debts []models.OptimizedDebt) error
	DeleteOptimizedDebtsByEventID(eventID int64) error
}


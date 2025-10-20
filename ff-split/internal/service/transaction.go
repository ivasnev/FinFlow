package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
)

// Transaction определяет методы для работы с транзакциями
type Transaction interface {
	GetTransactionsByEventID(ctx context.Context, eventID int64) ([]dto.TransactionResponse, error)
	GetTransactionByID(ctx context.Context, id int) (*dto.TransactionResponse, error)
	CreateTransaction(ctx context.Context, eventID int64, req *dto.TransactionRequest) (*dto.TransactionResponse, error)
	UpdateTransaction(ctx context.Context, id int, req *dto.TransactionRequest) (*dto.TransactionResponse, error)
	DeleteTransaction(ctx context.Context, id int) error
	GetDebtsByEventID(ctx context.Context, eventID int64, userID *int64) ([]dto.DebtDTO, error)
	GetDebtsByEventIDFromUser(eventID int64, userID int64) ([]dto.DebtDTO, error)
	GetDebtsByEventIDToUser(eventID int64, userID int64) ([]dto.DebtDTO, error)

	// Методы для работы с оптимизированными долгами
	OptimizeDebts(ctx context.Context, eventID int64) ([]dto.OptimizedDebtDTO, error)
	GetOptimizedDebtsByEventID(ctx context.Context, eventID int64, userID *int64) ([]dto.OptimizedDebtDTO, error)
	GetOptimizedDebtsByUserID(ctx context.Context, eventID, userID int64) ([]dto.OptimizedDebtDTO, error)
	GetOptimizedDebtsByEventIDFromUser(eventID int64, userID int64) ([]dto.OptimizedDebtDTO, error)
	GetOptimizedDebtsByEventIDToUser(eventID int64, userID int64) ([]dto.OptimizedDebtDTO, error)
}


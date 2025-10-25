package service

import (
	"context"
	"time"
)

// TransactionRequest представляет запрос на создание транзакции
type TransactionRequest struct {
	Type     string             `json:"type" binding:"required"`      // "percent" | "amount" | "units"
	FromUser int64              `json:"from_user" binding:"required"` // ID пользователя, который заплатил
	Amount   float64            `json:"amount" binding:"required"`    // Общая сумма
	Portion  map[string]float64 `json:"portion"`                      // Распределение (зависит от типа)
	Users    []int64            `json:"users" binding:"required"`     // Список пользователей-участников

	// Дополнительные поля для связи с сущностями
	Name                  string `json:"name" binding:"required"` // Название/описание транзакции
	TransactionCategoryID *int   `json:"transaction_category_id"` // ID категории транзакции
}

// TransactionResponse представляет ответ с информацией о транзакции
type TransactionResponse struct {
	ID                    int        `json:"id"`
	EventID               int64      `json:"event_id"`
	Name                  string     `json:"name"`
	TransactionCategoryID *int       `json:"transaction_category_id,omitempty"`
	Type                  string     `json:"type"`
	FromUser              int64      `json:"from_user"`
	Amount                float64    `json:"amount"`
	Datetime              time.Time  `json:"datetime"`
	Debts                 []DebtDTO  `json:"debts,omitempty"`
	Shares                []ShareDTO `json:"shares,omitempty"`
}

// TransactionListResponse представляет ответ со списком транзакций
type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
}

// DebtsUserResponse представляет пользователя в контексте долгов
type DebtsUserResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Photo string `json:"photo"`
}

// DebtDTO представляет информацию о долге
type DebtDTO struct {
	ID            int     `json:"id,omitempty"`
	FromUserID    int64   `json:"from_user_id"`
	ToUserID      int64   `json:"to_user_id"`
	Amount        float64 `json:"amount"`
	TransactionID int     `json:"transaction_id,omitempty"`

	FromUser  *DebtsUserResponse `json:"from_user,omitempty"`
	ToUser    *DebtsUserResponse `json:"to_user,omitempty"`
	Requestor *DebtsUserResponse `json:"requestor,omitempty"`
}

// DebtListResponse представляет ответ со списком долгов
type DebtListResponse []DebtDTO

// ShareDTO представляет информацию о доле в транзакции
type ShareDTO struct {
	ID            int     `json:"id,omitempty"`
	UserID        int64   `json:"user_id"`
	Value         float64 `json:"value"`
	TransactionID int     `json:"transaction_id,omitempty"`
}

// OptimizedDebtDTO представляет информацию об оптимизированных долгах
type OptimizedDebtDTO struct {
	ID         int     `json:"id,omitempty"`
	FromUserID int64   `json:"from_user_id"`
	ToUserID   int64   `json:"to_user_id"`
	Amount     float64 `json:"amount"`
	EventID    int64   `json:"event_id"`

	FromUser  *DebtsUserResponse `json:"from_user,omitempty"`
	ToUser    *DebtsUserResponse `json:"to_user,omitempty"`
	Requestor *DebtsUserResponse `json:"requestor,omitempty"`
}

// OptimizedDebtListResponse представляет ответ со списком оптимизированных долгов
type OptimizedDebtListResponse []OptimizedDebtDTO

// Transaction определяет методы для работы с транзакциями
type Transaction interface {
	GetTransactionsByEventID(ctx context.Context, eventID int64) ([]TransactionResponse, error)
	GetTransactionByID(ctx context.Context, id int) (*TransactionResponse, error)
	CreateTransaction(ctx context.Context, eventID int64, req *TransactionRequest) (*TransactionResponse, error)
	UpdateTransaction(ctx context.Context, id int, req *TransactionRequest) (*TransactionResponse, error)
	DeleteTransaction(ctx context.Context, id int) error
	GetDebtsByEventID(ctx context.Context, eventID int64, userID *int64) ([]DebtDTO, error)
	GetDebtsByEventIDFromUser(eventID int64, userID int64) ([]DebtDTO, error)
	GetDebtsByEventIDToUser(eventID int64, userID int64) ([]DebtDTO, error)

	// Методы для работы с оптимизированными долгами
	OptimizeDebts(ctx context.Context, eventID int64) ([]OptimizedDebtDTO, error)
	GetOptimizedDebtsByEventID(ctx context.Context, eventID int64, userID *int64) ([]OptimizedDebtDTO, error)
	GetOptimizedDebtsByUserID(ctx context.Context, eventID, userID int64) ([]OptimizedDebtDTO, error)
	GetOptimizedDebtsByEventIDFromUser(eventID int64, userID int64) ([]OptimizedDebtDTO, error)
	GetOptimizedDebtsByEventIDToUser(eventID int64, userID int64) ([]OptimizedDebtDTO, error)
}

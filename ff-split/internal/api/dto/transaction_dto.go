package dto

import "time"

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

type DebtsUserResponse struct {
	ID         int64  `json:"id"`
	ExternalID *int64 `json:"external_id"`
	Name       string `json:"name"`
	Photo      string `json:"photo"`
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
}

// OptimizedDebtListResponse представляет ответ со списком оптимизированных долгов
type OptimizedDebtListResponse []OptimizedDebtDTO

package models

import "time"

// DebtOptimizationState отличает актуальный, устаревший и ещё не рассчитанный результат.
type DebtOptimizationState struct {
	DeletedAt             *time.Time
	DebtRevision          int64
	OptimizedDebtRevision *int64
}

// Debt представляет долг одного пользователя другому
type Debt struct {
	// Settlement помечает обратное ребро возврата, которое не является расходом.
	Settlement    bool `gorm:"-"`
	ID            int
	TransactionID int
	FromUserID    int64
	ToUserID      int64
	Amount        float64

	// Отношения
	Transaction *Transaction
	FromUser    *User
	ToUser      *User
}

// OptimizedDebt представляет оптимизированные долги между пользователями
type OptimizedDebt struct {
	ID         int
	EventID    int64
	FromUserID int64
	ToUserID   int64
	Amount     float64
	CreatedAt  time.Time
	UpdatedAt  time.Time

	// Отношения
	FromUser *User
	ToUser   *User
}

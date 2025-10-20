package models

import "time"

// Transaction представляет транзакцию
type Transaction struct {
	ID                    int
	EventID               *int64
	Name                  string
	TransactionCategoryID *int
	Datetime              time.Time
	TotalPaid             float64
	PayerID               *int64
	SplitType             int

	// Отношения
	Event               *Event
	TransactionCategory *TransactionCategory
	Payer               *User
	Shares              []TransactionShare
	Debts               []Debt
}

// TransactionCategory представляет категорию транзакции
type TransactionCategory struct {
	ID     int
	Name   string
	IconID int

	// Отношения
	Icon         *Icon
	Transactions []Transaction
}

// TransactionShare представляет долю пользователя в транзакции
type TransactionShare struct {
	ID            int
	TransactionID int
	UserID        int64
	Value         float64

	// Отношения
	Transaction *Transaction
	User        *User
}

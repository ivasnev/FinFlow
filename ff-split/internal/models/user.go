package models

// User представляет модель пользователя
type User struct {
	ID              int64
	UserID          *int64
	NicknameCashed  string
	NameCashed      string
	PhotoUUIDCashed string
	IsDummy         bool

	// Отношения
	Events       []Event
	Activities   []Activity
	Transactions []Transaction
	Tasks        []Task
	Shares       []TransactionShare
	DebtsFrom    []Debt
	DebtsTo      []Debt
}

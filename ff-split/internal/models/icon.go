package models

// Icon представляет иконку для типа транзакции
type Icon struct {
	ID       int
	Name     string
	FileUUID string

	// Отношения
	TransactionCategories []TransactionCategory
}

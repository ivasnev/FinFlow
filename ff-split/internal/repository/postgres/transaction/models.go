package transaction

import "time"

// Transaction представляет транзакцию в БД
type Transaction struct {
	ID                    int       `gorm:"column:id;primaryKey;autoIncrement"`
	EventID               *int64    `gorm:"column:event_id"`
	Name                  string    `gorm:"column:name;not null"`
	TransactionCategoryID *int      `gorm:"column:transaction_category_id"`
	Datetime              time.Time `gorm:"column:datetime;default:CURRENT_TIMESTAMP"`
	TotalPaid             float64   `gorm:"column:total_paid;type:numeric(10,2);not null"`
	PayerID               *int64    `gorm:"column:payer_id"`
	SplitType             int       `gorm:"column:split_type;default:0;not null"`
}

// TableName задает имя таблицы для модели Transaction
func (Transaction) TableName() string {
	return "transactions"
}

// TransactionCategory представляет категорию транзакции в БД
type TransactionCategory struct {
	ID     int    `gorm:"column:id;primaryKey;autoIncrement"`
	Name   string `gorm:"column:name;not null"`
	IconID int    `gorm:"column:icon_id"`
}

// TableName задает имя таблицы для модели TransactionCategory
func (TransactionCategory) TableName() string {
	return "transaction_categories"
}

// TransactionShare представляет долю пользователя в транзакции в БД
type TransactionShare struct {
	ID            int     `gorm:"column:id;primaryKey;autoIncrement"`
	TransactionID int     `gorm:"column:transaction_id;uniqueIndex:uniq_tx_user"`
	UserID        int64   `gorm:"column:user_id;uniqueIndex:uniq_tx_user"`
	Value         float64 `gorm:"column:value;type:numeric(10,2);not null"`
}

// TableName задает имя таблицы для модели TransactionShare
func (TransactionShare) TableName() string {
	return "transaction_shares"
}

// Debt представляет долг одного пользователя другому в БД
type Debt struct {
	ID            int     `gorm:"column:id;primaryKey;autoIncrement"`
	TransactionID int     `gorm:"column:transaction_id;uniqueIndex:uniq_debt"`
	FromUserID    int64   `gorm:"column:from_user_id;uniqueIndex:uniq_debt"`
	ToUserID      int64   `gorm:"column:to_user_id;uniqueIndex:uniq_debt"`
	Amount        float64 `gorm:"column:amount;type:numeric(10,2);not null"`
}

// TableName задает имя таблицы для модели Debt
func (Debt) TableName() string {
	return "debts"
}

// OptimizedDebt представляет оптимизированные долги между пользователями в БД
type OptimizedDebt struct {
	ID         int       `gorm:"column:id;primaryKey;autoIncrement"`
	EventID    int64     `gorm:"column:event_id"`
	FromUserID int64     `gorm:"column:from_user_id"`
	ToUserID   int64     `gorm:"column:to_user_id"`
	Amount     float64   `gorm:"column:amount;type:numeric(10,2);not null"`
	CreatedAt  time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt  time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
}

// TableName задает имя таблицы для модели OptimizedDebt
func (OptimizedDebt) TableName() string {
	return "optimized_debts"
}

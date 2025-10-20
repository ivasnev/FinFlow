package category

// EventCategory представляет категорию мероприятия в БД
type EventCategory struct {
	ID     int    `gorm:"column:id;primaryKey;autoIncrement"`
	Name   string `gorm:"column:name;not null"`
	IconID int    `gorm:"column:icon_id"`
}

// TableName задает имя таблицы для модели EventCategory
func (EventCategory) TableName() string {
	return "event_categories"
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

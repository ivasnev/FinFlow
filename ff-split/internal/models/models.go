package models

import (
	"time"
)

// User представляет модель пользователя
type User struct {
	ID              int64  `gorm:"column:id;primaryKey;autoIncrement"`
	UserID          *int64 `gorm:"column:user_id;uniqueIndex"`
	NicknameCashed  string `gorm:"column:nickname_cashed"`
	NameCashed      string `gorm:"column:name_cashed"`
	PhotoUUIDCashed string `gorm:"column:photo_uuid_cashed"`
	IsDummy         bool   `gorm:"column:is_dummy;default:false"`

	// Отношения
	Events       []Event            `gorm:"many2many:user_event;foreignKey:UserID;references:ID"`
	Activities   []Activity         `gorm:"foreignKey:UserID;references:ID"`
	Transactions []Transaction      `gorm:"foreignKey:PayerID;references:UserID"`
	Tasks        []Task             `gorm:"foreignKey:UserID;references:ID"`
	Shares       []TransactionShare `gorm:"foreignKey:UserID;references:ID"`
	DebtsFrom    []Debt             `gorm:"foreignKey:FromUserID;references:UserID"`
	DebtsTo      []Debt             `gorm:"foreignKey:ToUserID;references:UserID"`
}

// TableName задает имя таблицы для модели User
func (User) TableName() string {
	return "users"
}

// EventCategory представляет категорию мероприятия
type EventCategory struct {
	ID     int    `gorm:"column:id;primaryKey;autoIncrement"`
	Name   string `gorm:"column:name;not null"`
	IconID int    `gorm:"column:icon_id"`

	// Отношения
	Icon   *Icon   `gorm:"foreignKey:IconID"`
	Events []Event `gorm:"foreignKey:CategoryID"`
}

// TableName задает имя таблицы для модели EventCategory
func (EventCategory) TableName() string {
	return "event_categories"
}

// Event представляет модель мероприятия
type Event struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string `gorm:"column:name;not null"`
	Description string `gorm:"column:description"`
	CategoryID  *int   `gorm:"column:category_id"`
	ImageID     string `gorm:"column:image_id"`
	Status      string `gorm:"column:status;default:active"`

	// Отношения
	Category     *EventCategory `gorm:"foreignKey:CategoryID"`
	Users        []User         `gorm:"many2many:user_event;foreignKey:ID;joinForeignKey:EventID;references:UserID;joinReferences:UserID"`
	Activities   []Activity     `gorm:"foreignKey:EventID"`
	Transactions []Transaction  `gorm:"foreignKey:EventID"`
	Tasks        []Task         `gorm:"foreignKey:EventID"`
}

// TableName задает имя таблицы для модели Event
func (Event) TableName() string {
	return "events"
}

// UserEvent представляет связь между пользователем и мероприятием
type UserEvent struct {
	UserID  int64 `gorm:"column:user_id;primaryKey"`
	EventID int64 `gorm:"column:event_id;primaryKey"`
}

// TableName задает имя таблицы для модели UserEvent
func (UserEvent) TableName() string {
	return "user_event"
}

// Activity представляет действие в системе
type Activity struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement"`
	EventID     *int64    `gorm:"column:event_id"`
	UserID      *int64    `gorm:"column:user_id"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`

	// Отношения
	Event *Event `gorm:"foreignKey:EventID"`
	User  *User  `gorm:"foreignKey:UserID;references:ID"`
}

// TableName задает имя таблицы для модели Activity
func (Activity) TableName() string {
	return "activities"
}

// Icon представляет иконку для типа транзакции
type Icon struct {
	ID       int    `gorm:"column:id;primaryKey"`
	Name     string `gorm:"column:name;not null"`
	FileUUID string `gorm:"column:file_uuid;not null"`

	// Отношения
	TransactionCategories []TransactionCategory `gorm:"foreignKey:IconID"`
}

// TableName задает имя таблицы для модели Icon
func (Icon) TableName() string {
	return "icons"
}

// TransactionCategory представляет категорию транзакции
type TransactionCategory struct {
	ID     int    `gorm:"column:id;primaryKey;autoIncrement"`
	Name   string `gorm:"column:name;not null"`
	IconID int    `gorm:"column:icon_id"`

	// Отношения
	Icon         *Icon         `gorm:"foreignKey:IconID"`
	Transactions []Transaction `gorm:"foreignKey:TransactionCategoryID"`
}

// TableName задает имя таблицы для модели TransactionCategory
func (TransactionCategory) TableName() string {
	return "transaction_categories"
}

// Transaction представляет транзакцию
type Transaction struct {
	ID                    int       `gorm:"column:id;primaryKey;autoIncrement"`
	EventID               *int64    `gorm:"column:event_id"`
	Name                  string    `gorm:"column:name;not null"`
	TransactionCategoryID *int      `gorm:"column:transaction_category_id"`
	Datetime              time.Time `gorm:"column:datetime;default:CURRENT_TIMESTAMP"`
	TotalPaid             float64   `gorm:"column:total_paid;type:numeric(10,2);not null"`
	PayerID               *int64    `gorm:"column:payer_id"`
	SplitType             int       `gorm:"column:split_type;default:0;not null"`

	// Отношения
	Event               *Event               `gorm:"foreignKey:EventID"`
	TransactionCategory *TransactionCategory `gorm:"foreignKey:TransactionCategoryID"`
	Payer               *User                `gorm:"foreignKey:PayerID;references:UserID"`
	Shares              []TransactionShare   `gorm:"foreignKey:TransactionID"`
	Debts               []Debt               `gorm:"foreignKey:TransactionID"`
}

// TableName задает имя таблицы для модели Transaction
func (Transaction) TableName() string {
	return "transactions"
}

// TransactionShare представляет долю пользователя в транзакции
type TransactionShare struct {
	ID            int     `gorm:"column:id;primaryKey;autoIncrement"`
	TransactionID int     `gorm:"column:transaction_id;uniqueIndex:uniq_tx_user"`
	UserID        int64   `gorm:"column:user_id;uniqueIndex:uniq_tx_user"`
	Value         float64 `gorm:"column:value;type:numeric(10,2);not null"`

	// Отношения
	Transaction *Transaction `gorm:"foreignKey:TransactionID"`
	User        *User        `gorm:"foreignKey:UserID;references:ID"`
}

// TableName задает имя таблицы для модели TransactionShare
func (TransactionShare) TableName() string {
	return "transaction_shares"
}

// Debt представляет долг одного пользователя другому
type Debt struct {
	ID            int     `gorm:"column:id;primaryKey;autoIncrement"`
	TransactionID int     `gorm:"column:transaction_id;uniqueIndex:uniq_debt"`
	FromUserID    int64   `gorm:"column:from_user_id;uniqueIndex:uniq_debt"`
	ToUserID      int64   `gorm:"column:to_user_id;uniqueIndex:uniq_debt"`
	Amount        float64 `gorm:"column:amount;type:numeric(10,2);not null"`

	// Отношения
	Transaction *Transaction `gorm:"foreignKey:TransactionID"`
	FromUser    *User        `gorm:"foreignKey:FromUserID;references:UserID"`
	ToUser      *User        `gorm:"foreignKey:ToUserID;references:UserID"`
}

// TableName задает имя таблицы для модели Debt
func (Debt) TableName() string {
	return "debts"
}

// Task представляет задачу в системе
type Task struct {
	ID          int       `gorm:"column:id;primaryKey;autoIncrement"`
	UserID      *int64    `gorm:"column:user_id"`
	EventID     *int64    `gorm:"column:event_id"`
	Title       string    `gorm:"column:title;not null"`
	Description string    `gorm:"column:description"`
	Priority    int       `gorm:"column:priority;default:0"`
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`

	// Отношения
	User  *User  `gorm:"foreignKey:UserID;references:ID"`
	Event *Event `gorm:"foreignKey:EventID"`
}

// TableName задает имя таблицы для модели Task
func (Task) TableName() string {
	return "tasks"
}

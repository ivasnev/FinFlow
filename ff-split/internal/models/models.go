package models

import (
	"time"
)

// User представляет пользователя в системе
type User struct {
	ID              int64  `json:"id" gorm:"primaryKey"`
	IDUser          int64  `json:"id_user" gorm:"uniqueIndex"`
	NicknameCashed  string `json:"nickname_cashed"`
	NameCashed      string `json:"name_cashed"`
	PhotoUUIDCashed string `json:"photo_uuid_cashed"`
	IsDummy         bool   `json:"is_dummy" gorm:"default:false"`
}

// Category представляет категорию для мероприятий
type Category struct {
	ID      int    `json:"id" gorm:"primaryKey"`
	Name    string `json:"name"`
	ImageID string `json:"image_id"`
}

// Event представляет мероприятие
type Event struct {
	ID          int64    `json:"id" gorm:"primaryKey"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CategoryID  int      `json:"category_id"`
	ImageID     string   `json:"image_id"`
	Status      string   `json:"status" gorm:"check:status IN ('active', 'archive')"`
	Category    Category `json:"-" gorm:"foreignKey:CategoryID"`
}

// UserEvent связывает пользователей и мероприятия
type UserEvent struct {
	IDUser  int64 `json:"id_user" gorm:"primaryKey"`
	IDEvent int64 `json:"id_event" gorm:"primaryKey"`
	User    User  `json:"-" gorm:"foreignKey:IDUser"`
	Event   Event `json:"-" gorm:"foreignKey:IDEvent"`
}

// Activity представляет событие в мероприятии
type Activity struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	IDEvent     int64     `json:"id_event"`
	IDUser      int64     `json:"id_user"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	Event       Event     `json:"-" gorm:"foreignKey:IDEvent"`
	User        User      `json:"-" gorm:"foreignKey:IDUser"`
}

// Task представляет задачу в мероприятии
type Task struct {
	ID          int       `json:"id" gorm:"primaryKey"`
	UserID      int64     `json:"user_id"`
	EventID     int64     `json:"event_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	User        User      `json:"-" gorm:"foreignKey:UserID"`
	Event       Event     `json:"-" gorm:"foreignKey:EventID"`
}

// TransactionType представляет тип транзакции
type TransactionType struct {
	ID     int    `json:"id" gorm:"primaryKey"`
	Name   string `json:"name"`
	IconID string `json:"icon_id"`
}

// Transaction представляет транзакцию
type Transaction struct {
	ID                int             `json:"id" gorm:"primaryKey"`
	EventID           int64           `json:"event_id"`
	Name              string          `json:"name"`
	TransactionTypeID int             `json:"transaction_type_id"`
	DateTime          time.Time       `json:"datetime"`
	TotalPaid         float64         `json:"total_paid"`
	PayerID           int64           `json:"payer_id"`
	Event             Event           `json:"-" gorm:"foreignKey:EventID"`
	TransactionType   TransactionType `json:"-" gorm:"foreignKey:TransactionTypeID"`
	Payer             User            `json:"-" gorm:"foreignKey:PayerID"`
}

// UserTransaction представляет участие пользователя в транзакции
type UserTransaction struct {
	ID            int         `json:"id" gorm:"primaryKey"`
	TransactionID int         `json:"transaction_id"`
	UserID        int64       `json:"user_id"`
	UserPart      float64     `json:"user_part"`
	Transaction   Transaction `json:"-" gorm:"foreignKey:TransactionID"`
	User          User        `json:"-" gorm:"foreignKey:UserID"`
}

// Icon представляет иконку
type Icon struct {
	ID       string `json:"id" gorm:"primaryKey"`
	Name     string `json:"name"`
	FileUUID string `json:"file_uuid"`
}

// EventMembersRequest представляет запрос на добавление участников в мероприятие
type EventMembersRequest struct {
	UserIDs      []int64  `json:"user_ids"`
	DummiesNames []string `json:"dummies_names"`
}

// EventTransactionTemporalResponse представляет запрос на получение данных о задолженностях
type EventTransactionTemporalResponse struct {
	TotalID   int `json:"total_id"`
	Requestor struct {
		Name  string `json:"name"`
		ID    int64  `json:"id"`
		Photo string `json:"photo"`
	} `json:"requestor"`
	Debtor struct {
		Name  string `json:"name"`
		ID    int64  `json:"id"`
		Photo string `json:"photo"`
	} `json:"debtor"`
	Amount float64 `json:"amount"`
}

// EventRequest представляет запрос на создание мероприятия
type EventRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	CategoryID  int                 `json:"category_id"`
	Members     EventMembersRequest `json:"members"`
}

// EventResponse представляет ответ с информацией о мероприятии
type EventResponse struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	CategoryID int    `json:"category_id"`
	PhotoID    string `json:"photo_id"`
	Balance    int    `json:"balance,omitempty"`
}

// CategoryResponse представляет ответ с информацией о категории
type CategoryResponse struct {
	ID     int    `json:"id"`
	IconID int    `json:"icon_id"`
	Name   string `json:"name"`
}

// ActivityResponse представляет ответ с информацией о событии в мероприятии
type ActivityResponse struct {
	ActivityID  int       `json:"activity_id"`
	Description string    `json:"description"`
	IconID      string    `json:"icon_id"`
	DateTime    time.Time `json:"datetime"`
}

// TransactionResponse представляет ответ с информацией о транзакции
type TransactionResponse struct {
	TransactionID     int    `json:"transaction_id"`
	Name              string `json:"name"`
	TransactionTypeID struct {
		ID     int    `json:"id"`
		IconID string `json:"icon_id"`
	} `json:"transaction_type_id"`
	DateTime  time.Time `json:"datetime"`
	UserPart  float64   `json:"user_part"`
	TotalPaid float64   `json:"total_paid"`
	PayerName string    `json:"payer_name"`
}

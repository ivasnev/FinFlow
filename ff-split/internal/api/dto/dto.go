package dto

import "time"

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

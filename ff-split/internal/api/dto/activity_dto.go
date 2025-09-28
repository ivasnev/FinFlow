package dto

import "time"

// ActivityRequest представляет DTO для запроса создания/обновления активности
type ActivityRequest struct {
	UserID      *int64 `json:"user_id"`
	Description string `json:"description" binding:"required"`
}

// ActivityResponse представляет DTO для ответа с данными активности
type ActivityResponse struct {
	ID          int       `json:"id"`
	EventID     *int64    `json:"event_id,omitempty"`
	UserID      *int64    `json:"user_id,omitempty"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ActivityListResponse представляет DTO для ответа со списком активностей
type ActivityListResponse struct {
	Activities []ActivityResponse `json:"activities"`
}

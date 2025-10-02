package dto

import "time"

// ActivityRequest представляет DTO для запроса создания/обновления активности
type ActivityRequest struct {
	UserID      *int64 `json:"user_id"`
	Description string `json:"description" binding:"required"`
	IconID      int    `json:"icon_id"`
}

// ActivityResponse представляет DTO для ответа с данными активности
type ActivityResponse struct {
	ActivityID  int       `json:"activity_id"`
	Description string    `json:"description"`
	IconID      int       `json:"icon_id"`
	Datetime    time.Time `json:"datetime"`
}

// ActivityListResponse представляет DTO для ответа со списком активностей
type ActivityListResponse struct {
	Activities []ActivityResponse `json:"activities"`
}

package dto

// EventRequest представляет DTO для запроса создания/обновления мероприятия
type EventRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	CategoryID  *int   `json:"category_id"`
	ImageID     string `json:"image_id"`
	Status      string `json:"status"`
}

// EventResponse представляет DTO для ответа с данными мероприятия
type EventResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CategoryID  *int   `json:"category_id,omitempty"`
	ImageID     string `json:"image_id,omitempty"`
	Status      string `json:"status"`
}

// EventListResponse представляет DTO для ответа со списком мероприятий
type EventListResponse struct {
	Events []EventResponse `json:"events"`
}

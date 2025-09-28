package dto

// EventRequest представляет DTO для запроса создания/обновления мероприятия
type EventRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	CategoryID  *int            `json:"category_id,omitempty"`
	Members     EventMembersDTO `json:"members"`
}

// EventMembersDTO представляет DTO для передачи данных о членах мероприятия
type EventMembersDTO struct {
	UserIDs      []int64  `json:"user_ids"`
	DummiesNames []string `json:"dummies_names"`
}

// EventResponse представляет DTO для ответа с данными мероприятия
type EventResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	CategoryID  *int   `json:"category_id,omitempty"`
	PhotoID     string `json:"photo_id,omitempty"`
	Balance     *int   `json:"balance,omitempty"`
}

// EventListResponse представляет DTO для ответа со списком мероприятий
type EventListResponse struct {
	Events []EventResponse `json:"events"`
}

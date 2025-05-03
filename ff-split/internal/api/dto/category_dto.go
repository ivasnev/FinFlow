package dto

// CategoryRequest представляет DTO для запроса создания/обновления категории
type CategoryRequest struct {
	Name   string `json:"name" binding:"required"`
	IconID string `json:"icon_id"`
}

// CategoryResponse представляет DTO для ответа с данными категории
type CategoryResponse struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	IconID string `json:"icon_id,omitempty"`
}

// CategoryListResponse представляет DTO для ответа со списком категорий
type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

// CategoryTypesResponse представляет DTO для ответа со списком типов категорий
type CategoryTypesResponse struct {
	Types []string `json:"types"`
}

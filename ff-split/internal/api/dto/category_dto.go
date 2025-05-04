package dto

// CategoryRequest представляет DTO для запроса создания/обновления категории
type CategoryRequest struct {
	Name   string `json:"name" binding:"required"`
	IconID int    `json:"icon_id"`
}

// CategoryDTO представляет DTO для передачи данных категорий
type CategoryDTO struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	IconID int    `json:"icon_id"`

	Icon IconDTO `json:"icon"`
}

type IconDTO struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ExternalUuid string `json:"external_uuid"`
}

// CategoryResponse представляет DTO для ответа с данными категории
type CategoryResponse struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`
	Icon IconDTO `json:"icon"`
}

// CategoryListResponse представляет DTO для ответа со списком категорий
type CategoryListResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

// CategoryTypesResponse представляет DTO для ответа со списком типов категорий
type CategoryTypesResponse []string

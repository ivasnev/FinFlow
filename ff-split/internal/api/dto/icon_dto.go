package dto

// IconFullDTO представляет полное DTO для иконки
type IconFullDTO struct {
	ID       uint   `json:"id"`
	Name     string `json:"name" binding:"required"`
	FileUUID string `json:"file_uuid" binding:"required"`
}

// IconResponse представляет ответ на операцию с иконкой
type IconResponse IconFullDTO

// IconListResponse представляет ответ со списком иконок
type IconListResponse []IconFullDTO

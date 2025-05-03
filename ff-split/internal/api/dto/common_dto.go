package dto

// ErrorResponse представляет DTO для ошибок
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse представляет DTO для успешного ответа без данных
type SuccessResponse struct {
	Success bool `json:"success"`
}

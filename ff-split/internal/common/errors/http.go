package errors

import "net/http"

const (
	ErrCodeNotFound      = "not_found"
	ErrCodeAlreadyExists = "already_exists"
	ErrCodeValidation    = "validation"
	ErrCodeLogic         = "error_logic"
	ErrCodeDatabase      = "error_database"
	ErrCodeInternal      = "error_internal"
)

type ErrorResponse struct {
	ID    string              `json:"id"`
	Error ErrorResponseDetail `json:"error"`
}

type ErrorResponseDetail struct {
	Data *ErrorResponseAdditionalData `json:"data,omitempty"`

	Code    string `json:"code" example:"validation"`
	Message string `json:"message" example:"invalid input"`
}

type ErrorResponseAdditionalData struct {
	Slug string `json:"slug" example:"slug"`
}

func NewErrorResponse(r *http.Request, code string, message string) *ErrorResponse {
	return &ErrorResponse{
		ID: r.Header.Get("X-Request-ID"),
		Error: ErrorResponseDetail{
			Code:    code,
			Message: message,
		},
	}
}
func NewValidationErrorResponse(r *http.Request, message string) *ErrorResponse {
	return NewErrorResponse(r, ErrCodeValidation, message)
}

func NewAlreadyExistsErrorResponse(r *http.Request, message string) *ErrorResponse {
	return NewErrorResponse(r, ErrCodeAlreadyExists, message)
}

func NewNotFoundErrorResponse(r *http.Request, message string) *ErrorResponse {
	return NewErrorResponse(r, ErrCodeNotFound, message)
}

func NewLogicErrorResponse(r *http.Request, message string, slug string) *ErrorResponse {
	errorResponse := NewErrorResponse(r, ErrCodeLogic, message)
	if slug != "" {
		errorResponse.Error.Data = &ErrorResponseAdditionalData{
			Slug: slug,
		}
	}
	return errorResponse
}

func NewDatabaseErrorResponse(r *http.Request, message string) *ErrorResponse {
	return NewErrorResponse(r, ErrCodeDatabase, message)
}

func NewInternalErrorResponse(r *http.Request, message string) *ErrorResponse {
	return NewErrorResponse(r, ErrCodeInternal, message)
}

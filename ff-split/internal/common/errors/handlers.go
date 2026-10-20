package errors

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"gorm.io/gorm"
)

func isDatabaseError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, sql.ErrNoRows) ||
		errors.Is(err, sql.ErrTxDone) ||
		errors.Is(err, sql.ErrConnDone) {
		return true
	}

	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr)
}

// HTTPErrorHandler переводит доменные ошибки в HTTP-ответы.
func HTTPErrorHandler(c *gin.Context, err error) {
	var errorResponse *ErrorResponse
	var code int

	var validationError *ValidationError
	var alreadyExistsError *AlreadyExistsError
	var entityNotFoundError *EntityNotFoundError
	var logicError *LogicError

	switch {
	case errors.Is(err, ErrInvitationUnavailable):
		errorResponse = NewErrorResponse(c.Request, "invitation_unavailable", err.Error())
		code = http.StatusGone
	case errors.Is(err, ErrForbidden):
		errorResponse = NewErrorResponse(c.Request, "forbidden", err.Error())
		code = http.StatusForbidden
	case errors.Is(err, ErrOptimizationBusy):
		c.Header("Retry-After", "1")
		errorResponse = NewErrorResponse(c.Request, "busy", err.Error())
		code = http.StatusTooManyRequests
	case errors.Is(err, ErrDebtRevisionChanged):
		errorResponse = NewErrorResponse(c.Request, "conflict", err.Error())
		code = http.StatusConflict
	case errors.As(err, &validationError):
		errorResponse = NewValidationErrorResponse(c.Request, validationError.Message)
		code = http.StatusBadRequest
	case errors.As(err, &alreadyExistsError):
		errorResponse = NewAlreadyExistsErrorResponse(c.Request, alreadyExistsError.Error())
		code = http.StatusConflict
	case errors.As(err, &entityNotFoundError):
		errorResponse = NewNotFoundErrorResponse(c.Request, entityNotFoundError.Error())
		code = http.StatusNotFound
	case errors.Is(err, gorm.ErrRecordNotFound):
		errorResponse = NewNotFoundErrorResponse(c.Request, "запись не найдена")
		code = http.StatusNotFound
	case errors.As(err, &logicError):
		errorResponse = NewLogicErrorResponse(c.Request, logicError.Error(), "")
		code = http.StatusBadRequest
	case isDatabaseError(err):
		errorResponse = NewDatabaseErrorResponse(c.Request, err.Error())
		code = http.StatusInternalServerError
	default:
		errorResponse = NewInternalErrorResponse(c.Request, err.Error())
		code = http.StatusInternalServerError
	}

	c.JSON(code, errorResponse)
}

package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// TransactionHandler обработчик для работы с транзакциями
type TransactionHandler struct {
	service service.TransactionServiceInterface
}

// NewTransactionHandler создает новый обработчик для работы с транзакциями
func NewTransactionHandler(service service.TransactionServiceInterface) *TransactionHandler {
	return &TransactionHandler{service: service}
}

// GetTransactionsByEventID возвращает список транзакций мероприятия
// @Summary Получить все транзакции мероприятия
// @Description Возвращает список всех транзакций, связанных с указанным мероприятием
// @Tags транзакции
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Success 200 {object} dto.TransactionListResponse "Список транзакций"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат ID мероприятия"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/transaction [get]
func (h *TransactionHandler) GetTransactionsByEventID(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_event", "неверный формат ID мероприятия"))
		return
	}

	transactions, err := h.service.GetTransactionsByEventID(c.Request.Context(), eventID)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении транзакций: %w", err))
		return
	}

	c.JSON(http.StatusOK, dto.TransactionListResponse{Transactions: transactions})
}

// GetTransactionByID возвращает транзакцию по ID
// @Summary Получить транзакцию по ID
// @Description Возвращает информацию о конкретной транзакции по её ID
// @Tags транзакции
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id_transaction path int true "ID транзакции"
// @Success 200 {object} dto.TransactionResponse "Информация о транзакции"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат ID"
// @Failure 404 {object} errors.ErrorResponse "Транзакция не найдена"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/transaction/{id_transaction} [get]
func (h *TransactionHandler) GetTransactionByID(c *gin.Context) {
	idStr := c.Param("id_transaction")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_transaction", "неверный формат ID транзакции"))
		return
	}

	transaction, err := h.service.GetTransactionByID(c.Request.Context(), id)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении транзакции: %w", err))
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// CreateTransaction создает новую транзакцию
// @Summary Создать новую транзакцию
// @Description Создает новую транзакцию в рамках указанного мероприятия
// @Tags транзакции
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param transaction body dto.TransactionRequest true "Данные транзакции"
// @Success 201 {object} dto.TransactionResponse "Созданная транзакция"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат данных запроса"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/transaction [post]
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_event", "неверный формат ID мероприятия"))
		return
	}

	var request dto.TransactionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("request_body", err.Error()))
		return
	}

	transaction, err := h.service.CreateTransaction(c.Request.Context(), eventID, &request)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при создании транзакции: %w", err))
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

// UpdateTransaction обновляет существующую транзакцию
// @Summary Обновить транзакцию
// @Description Обновляет существующую транзакцию по ID
// @Tags транзакции
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id_transaction path int true "ID транзакции"
// @Param transaction body dto.TransactionRequest true "Данные транзакции"
// @Success 200 {object} dto.TransactionResponse "Обновленная транзакция"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат данных запроса"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/transaction/{id_transaction} [put]
func (h *TransactionHandler) UpdateTransaction(c *gin.Context) {
	idStr := c.Param("id_transaction")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_transaction", "неверный формат ID транзакции"))
		return
	}

	var request dto.TransactionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("request_body", err.Error()))
		return
	}

	transaction, err := h.service.UpdateTransaction(c.Request.Context(), id, &request)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при обновлении транзакции: %w", err))
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// DeleteTransaction удаляет транзакцию
// @Summary Удалить транзакцию
// @Description Удаляет транзакцию по ID
// @Tags транзакции
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id_transaction path int true "ID транзакции"
// @Success 204 "Транзакция успешно удалена"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат ID"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/transaction/{id_transaction} [delete]
func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	idStr := c.Param("id_transaction")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_transaction", "неверный формат ID транзакции"))
		return
	}

	err = h.service.DeleteTransaction(c.Request.Context(), id)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при удалении транзакции: %w", err))
		return
	}

	c.Status(http.StatusNoContent)
}

// GetDebtsByEventID возвращает долги мероприятия
// @Summary Получить долги мероприятия
// @Description Возвращает все долги, связанные с мероприятием
// @Tags долги
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Success 200 {array} dto.DebtDTO "Список долгов"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат ID мероприятия"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/debts [get]
func (h *TransactionHandler) GetDebtsByEventID(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_event", "неверный формат ID мероприятия"))
		return
	}
	var userID *int64

	if id, exist := c.Get("user_id"); exist {
		if idInt, parsed := id.(int64); parsed {
			userID = &idInt
		}
	}

	debts, err := h.service.GetDebtsByEventID(c.Request.Context(), eventID, userID)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении долгов: %w", err))
		return
	}

	c.JSON(http.StatusOK, dto.DebtListResponse(debts))
}

// OptimizeDebts оптимизирует долги мероприятия
// @Summary Оптимизировать долги мероприятия
// @Description Запускает алгоритм оптимизации долгов и возвращает оптимизированный список долгов
// @Tags долги, оптимизация
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Success 200 {array} dto.OptimizedDebtDTO "Список оптимизированных долгов"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат ID мероприятия"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/optimized-debts [post]
func (h *TransactionHandler) OptimizeDebts(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_event", "неверный формат ID мероприятия"))
		return
	}

	optimizedDebts, err := h.service.OptimizeDebts(c.Request.Context(), eventID)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при оптимизации долгов: %w", err))
		return
	}

	c.JSON(http.StatusOK, dto.OptimizedDebtListResponse(optimizedDebts))
}

// GetOptimizedDebtsByEventID возвращает оптимизированные долги мероприятия
// @Summary Получить оптимизированные долги мероприятия
// @Description Возвращает список оптимизированных долгов мероприятия
// @Tags долги, оптимизация
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Success 200 {array} dto.OptimizedDebtDTO "Список оптимизированных долгов"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат ID мероприятия"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/optimized-debts [get]
func (h *TransactionHandler) GetOptimizedDebtsByEventID(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_event", "неверный формат ID мероприятия"))
		return
	}

	optimizedDebts, err := h.service.GetOptimizedDebtsByEventID(c.Request.Context(), eventID)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении оптимизированных долгов: %w", err))
		return
	}

	c.JSON(http.StatusOK, dto.OptimizedDebtListResponse(optimizedDebts))
}

// GetOptimizedDebtsByUserID возвращает оптимизированные долги пользователя в мероприятии
// @Summary Получить оптимизированные долги пользователя
// @Description Возвращает список оптимизированных долгов для конкретного пользователя в мероприятии
// @Tags долги, оптимизация
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id_user path int true "ID пользователя"
// @Success 200 {array} dto.OptimizedDebtDTO "Список оптимизированных долгов пользователя"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат ID"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/user/{id_user}/optimized-debts [get]
func (h *TransactionHandler) GetOptimizedDebtsByUserID(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_event", "неверный формат ID мероприятия"))
		return
	}

	userID, err := strconv.ParseInt(c.Param("id_user"), 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_user", "неверный формат ID пользователя"))
		return
	}

	optimizedDebts, err := h.service.GetOptimizedDebtsByUserID(c.Request.Context(), eventID, userID)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении оптимизированных долгов: %w", err))
		return
	}

	c.JSON(http.StatusOK, dto.OptimizedDebtListResponse(optimizedDebts))
}

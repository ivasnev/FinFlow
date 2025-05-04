package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
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
func (h *TransactionHandler) GetTransactionsByEventID(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID мероприятия"})
		return
	}

	transactions, err := h.service.GetTransactionsByEventID(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.TransactionListResponse{Transactions: transactions})
}

// GetTransactionByID возвращает транзакцию по ID
func (h *TransactionHandler) GetTransactionByID(c *gin.Context) {
	idStr := c.Param("id_transaction")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID транзакции"})
		return
	}

	transaction, err := h.service.GetTransactionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// CreateTransaction создает новую транзакцию
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID мероприятия"})
		return
	}

	var request dto.TransactionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transaction, err := h.service.CreateTransaction(c.Request.Context(), eventID, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}

// UpdateTransaction обновляет существующую транзакцию
func (h *TransactionHandler) UpdateTransaction(c *gin.Context) {
	idStr := c.Param("id_transaction")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID транзакции"})
		return
	}

	var request dto.TransactionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transaction, err := h.service.UpdateTransaction(c.Request.Context(), id, &request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// DeleteTransaction удаляет транзакцию
func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	idStr := c.Param("id_transaction")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID транзакции"})
		return
	}

	err = h.service.DeleteTransaction(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetDebtsByEventID возвращает долги мероприятия
func (h *TransactionHandler) GetDebtsByEventID(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID мероприятия"})
		return
	}

	debts, err := h.service.GetDebtsByEventID(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.DebtListResponse(debts))
}

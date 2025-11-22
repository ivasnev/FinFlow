package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
)

// GetTransactionsByEventID возвращает список транзакций мероприятия
func (s *ServerHandler) GetTransactionsByEventID(c *gin.Context, idEvent int64) {
	transactions, err := s.transactionService.GetTransactionsByEventID(c.Request.Context(), idEvent)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении транзакций: %w", err))
		return
	}

	// Конвертируем DTO в API типы
	apiTransactions := make([]api.TransactionResponse, 0, len(transactions))
	for _, t := range transactions {
		apiTransactions = append(apiTransactions, convertTransactionToAPI(&t))
	}

	c.JSON(http.StatusOK, api.TransactionListResponse{Transactions: &apiTransactions})
}

// GetTransactionByID возвращает транзакцию по ID
func (s *ServerHandler) GetTransactionByID(c *gin.Context, idEvent int64, idTransaction int) {
	transaction, err := s.transactionService.GetTransactionByID(c.Request.Context(), idTransaction)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении транзакции: %w", err))
		return
	}

	c.JSON(http.StatusOK, convertTransactionToAPI(transaction))
}

// CreateTransaction создает новую транзакцию
func (s *ServerHandler) CreateTransaction(c *gin.Context, idEvent int64) {
	var apiRequest api.TransactionRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
	c.JSON(http.StatusBadRequest, api.ErrorResponse{
		Id: c.GetHeader("X-Request-ID"),
		Error: api.ErrorResponseDetail{
			Code:    "validation",
			Message: "некорректные данные запроса",
		},
	})
		return
	}

	// Конвертируем API типы в DTO
	dtoRequest := convertTransactionRequestToDTO(&apiRequest)

	transaction, err := s.transactionService.CreateTransaction(c.Request.Context(), idEvent, &dtoRequest)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при создании транзакции: %w", err))
		return
	}

	c.JSON(http.StatusCreated, convertTransactionToAPI(transaction))
}

// UpdateTransaction обновляет существующую транзакцию
func (s *ServerHandler) UpdateTransaction(c *gin.Context, idEvent int64, idTransaction int) {
	var apiRequest api.TransactionRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
	c.JSON(http.StatusBadRequest, api.ErrorResponse{
		Id: c.GetHeader("X-Request-ID"),
		Error: api.ErrorResponseDetail{
			Code:    "validation",
			Message: "некорректные данные запроса",
		},
	})
		return
	}

	// Конвертируем API типы в DTO
	dtoRequest := convertTransactionRequestToDTO(&apiRequest)

	transaction, err := s.transactionService.UpdateTransaction(c.Request.Context(), idTransaction, &dtoRequest)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при обновлении транзакции: %w", err))
		return
	}

	c.JSON(http.StatusOK, convertTransactionToAPI(transaction))
}

// DeleteTransaction удаляет транзакцию
func (s *ServerHandler) DeleteTransaction(c *gin.Context, idEvent int64, idTransaction int) {
	err := s.transactionService.DeleteTransaction(c.Request.Context(), idTransaction)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при удалении транзакции: %w", err))
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{Success: true})
}

// GetDebtsByEventID возвращает список долгов мероприятия
func (s *ServerHandler) GetDebtsByEventID(c *gin.Context, idEvent int64) {
	debts, err := s.transactionService.GetDebtsByEventID(c.Request.Context(), idEvent, nil)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении долгов: %w", err))
		return
	}

	// Конвертируем DTO в API типы
	apiDebts := make([]api.DebtDTO, 0, len(debts))
	for _, d := range debts {
		apiDebts = append(apiDebts, convertDebtToAPI(&d))
	}

	c.JSON(http.StatusOK, api.DebtListResponse{Debts: &apiDebts})
}

// GetOptimizedDebtsByEventID возвращает оптимизированные долги
func (s *ServerHandler) GetOptimizedDebtsByEventID(c *gin.Context, idEvent int64) {
	debts, err := s.transactionService.GetOptimizedDebtsByEventID(c.Request.Context(), idEvent, nil)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении оптимизированных долгов: %w", err))
		return
	}

	// Конвертируем DTO в API типы
	apiDebts := make([]api.OptimizedDebtDTO, 0, len(debts))
	for _, d := range debts {
		apiDebts = append(apiDebts, convertOptimizedDebtToAPI(&d))
	}

	c.JSON(http.StatusOK, api.OptimizedDebtListResponse{OptimizedDebts: &apiDebts})
}

// OptimizeDebts оптимизирует долги мероприятия
func (s *ServerHandler) OptimizeDebts(c *gin.Context, idEvent int64) {
	debts, err := s.transactionService.OptimizeDebts(c.Request.Context(), idEvent)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при оптимизации долгов: %w", err))
		return
	}

	// Конвертируем DTO в API типы
	apiDebts := make([]api.OptimizedDebtDTO, 0, len(debts))
	for _, d := range debts {
		apiDebts = append(apiDebts, convertOptimizedDebtToAPI(&d))
	}

	c.JSON(http.StatusOK, api.OptimizedDebtListResponse{OptimizedDebts: &apiDebts})
}

// GetOptimizedDebtsByUserID возвращает оптимизированные долги пользователя
func (s *ServerHandler) GetOptimizedDebtsByUserID(c *gin.Context, idEvent int64, idUser int64) {
	debts, err := s.transactionService.GetOptimizedDebtsByUserID(c.Request.Context(), idEvent, idUser)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении оптимизированных долгов пользователя: %w", err))
		return
	}

	// Конвертируем DTO в API типы
	apiDebts := make([]api.OptimizedDebtDTO, 0, len(debts))
	for _, d := range debts {
		apiDebts = append(apiDebts, convertOptimizedDebtToAPI(&d))
	}

	c.JSON(http.StatusOK, api.OptimizedDebtListResponse{OptimizedDebts: &apiDebts})
}

// Helper functions для конвертации типов

func convertTransactionToAPI(t *service.TransactionResponse) api.TransactionResponse {
	// Конвертируем shares
	var shares *[]api.ShareDTO
	if len(t.Shares) > 0 {
		apiShares := make([]api.ShareDTO, 0, len(t.Shares))
		for _, s := range t.Shares {
			apiShares = append(apiShares, api.ShareDTO{
				Id:            &s.ID,
				TransactionId: &s.TransactionID,
				UserId:        &s.UserID,
				Value:         &s.Value,
			})
		}
		shares = &apiShares
	}

	// Конвертируем debts
	var debts *[]api.DebtDTO
	if len(t.Debts) > 0 {
		apiDebts := make([]api.DebtDTO, 0, len(t.Debts))
		for _, d := range t.Debts {
			apiDebts = append(apiDebts, convertDebtToAPI(&d))
		}
		debts = &apiDebts
	}

	return api.TransactionResponse{
		Id:                    &t.ID,
		EventId:               &t.EventID,
		Name:                  &t.Name,
		Amount:                &t.Amount,
		FromUser:              &t.FromUser,
		Type:                  &t.Type,
		TransactionCategoryId: t.TransactionCategoryID,
		Datetime:              &t.Datetime,
		Shares:                shares,
		Debts:                 debts,
	}
}

func convertTransactionRequestToDTO(req *api.TransactionRequest) service.TransactionRequest {
	dtoReq := service.TransactionRequest{
		Name:     req.Name,
		Amount:   req.Amount,
		FromUser: req.FromUser,
		Type:     string(req.Type),
	}

	if req.Users != nil {
		dtoReq.Users = req.Users
	}

	if req.Portion != nil {
		dtoReq.Portion = *req.Portion
	}

	if req.TransactionCategoryId != nil {
		dtoReq.TransactionCategoryID = req.TransactionCategoryId
	}

	return dtoReq
}

func convertDebtToAPI(d *service.DebtDTO) api.DebtDTO {
	return api.DebtDTO{
		Id:            &d.ID,
		TransactionId: &d.TransactionID,
		FromUserId:    &d.FromUserID,
		ToUserId:      &d.ToUserID,
		Amount:        &d.Amount,
	}
}

func convertOptimizedDebtToAPI(d *service.OptimizedDebtDTO) api.OptimizedDebtDTO {
	return api.OptimizedDebtDTO{
		Id:         &d.ID,
		EventId:    &d.EventID,
		FromUserId: &d.FromUserID,
		ToUserId:   &d.ToUserID,
		Amount:     &d.Amount,
	}
}

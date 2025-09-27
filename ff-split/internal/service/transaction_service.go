package service

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"math"
	"strconv"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
)

// TransactionServiceImpl реализация сервиса транзакций
type TransactionServiceImpl struct {
	transactionRepo repository.TransactionRepository
	eventRepo       repository.EventRepository
}

// NewTransactionService создает новый экземпляр сервиса транзакций
func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	eventRepo repository.EventRepository,
) *TransactionServiceImpl {
	return &TransactionServiceImpl{
		transactionRepo: transactionRepo,
		eventRepo:       eventRepo,
	}
}

// GetTransactionsByEventID возвращает все транзакции по ID мероприятия
func (s *TransactionServiceImpl) GetTransactionsByEventID(ctx context.Context, eventID int64) ([]dto.TransactionResponse, error) {
	transactions, err := s.transactionRepo.GetTransactionsByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	var responses []dto.TransactionResponse
	for _, transaction := range transactions {
		// Получаем данные о пользователях транзакции
		userTransactions, err := s.transactionRepo.GetTransactionUsers(ctx, transaction.ID)
		if err != nil {
			return nil, err
		}

		// Для каждого пользователя создаем отдельный ответ
		for _, userTransaction := range userTransactions {
			response := mapTransactionToResponse(transaction)
			response.UserPart = userTransaction.UserPart
			responses = append(responses, response)
		}
	}

	return responses, nil
}

// GetTransactionByID возвращает транзакцию по ID
func (s *TransactionServiceImpl) GetTransactionByID(ctx context.Context, id int) (dto.TransactionResponse, error) {
	transaction, err := s.transactionRepo.GetTransactionByID(ctx, id)
	if err != nil {
		return dto.TransactionResponse{}, err
	}

	return mapTransactionToResponse(transaction), nil
}

// CreateTransaction создает новую транзакцию
func (s *TransactionServiceImpl) CreateTransaction(ctx context.Context, transaction models.Transaction, userIDs []int64) (dto.TransactionResponse, error) {
	// Создаем транзакцию
	createdTransaction, err := s.transactionRepo.CreateTransaction(ctx, transaction)
	if err != nil {
		return dto.TransactionResponse{}, err
	}

	// Распределяем оплату равномерно между участниками
	if len(userIDs) > 0 {
		// Вычисляем долю каждого пользователя
		userPart := transaction.TotalPaid / float64(len(userIDs))
		userPart = math.Round(userPart*100) / 100 // Округляем до 2 знаков после запятой

		// Добавляем пользователей в транзакцию
		for _, userID := range userIDs {
			userTransaction := models.UserTransaction{
				TransactionID: createdTransaction.ID,
				UserID:        userID,
				UserPart:      userPart,
			}

			if err := s.transactionRepo.AddUserToTransaction(ctx, userTransaction); err != nil {
				return dto.TransactionResponse{}, err
			}
		}
	}

	return mapTransactionToResponse(createdTransaction), nil
}

// UpdateTransaction обновляет существующую транзакцию
func (s *TransactionServiceImpl) UpdateTransaction(ctx context.Context, transaction models.Transaction, userIDs []int64) error {
	// Обновляем транзакцию
	if err := s.transactionRepo.UpdateTransaction(ctx, transaction); err != nil {
		return err
	}

	// Удаляем всех текущих пользователей транзакции
	userTransactions, err := s.transactionRepo.GetTransactionUsers(ctx, transaction.ID)
	if err != nil {
		return err
	}

	for _, userTransaction := range userTransactions {
		if err := s.transactionRepo.RemoveUserFromTransaction(ctx, userTransaction.UserID, transaction.ID); err != nil {
			return err
		}
	}

	// Добавляем новых пользователей
	if len(userIDs) > 0 {
		// Вычисляем долю каждого пользователя
		userPart := transaction.TotalPaid / float64(len(userIDs))
		userPart = math.Round(userPart*100) / 100 // Округляем до 2 знаков после запятой

		// Добавляем пользователей в транзакцию
		for _, userID := range userIDs {
			userTransaction := models.UserTransaction{
				TransactionID: transaction.ID,
				UserID:        userID,
				UserPart:      userPart,
			}

			if err := s.transactionRepo.AddUserToTransaction(ctx, userTransaction); err != nil {
				return err
			}
		}
	}

	return nil
}

// DeleteTransaction удаляет транзакцию
func (s *TransactionServiceImpl) DeleteTransaction(ctx context.Context, id int) error {
	return s.transactionRepo.DeleteTransaction(ctx, id)
}

// GetTemporalTransactionsByEventID возвращает оптимизированные данные о задолженностях
func (s *TransactionServiceImpl) GetTemporalTransactionsByEventID(ctx context.Context, eventID int64) ([]dto.EventTransactionTemporalResponse, error) {
	// Получаем все транзакции для мероприятия
	transactions, err := s.transactionRepo.GetTransactionsByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Получаем всех пользователей мероприятия
	users, err := s.eventRepo.GetEventUsers(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Создаем карту долгов пользователей
	debts := make(map[string]map[string]int)
	userMap := make(map[int64]models.User)

	// Инициализируем карту для каждого пользователя
	for _, user := range users {
		userKey := strconv.FormatInt(user.IDUser, 10)
		debts[userKey] = make(map[string]int)
		userMap[user.IDUser] = user
	}

	// Обрабатываем все транзакции
	for _, transaction := range transactions {
		// Получаем участников транзакции
		userTransactions, err := s.transactionRepo.GetTransactionUsers(ctx, transaction.ID)
		if err != nil {
			return nil, err
		}

		payerKey := strconv.FormatInt(transaction.PayerID, 10)

		// Добавляем долги для каждого участника
		for _, userTransaction := range userTransactions {
			debtorKey := strconv.FormatInt(userTransaction.UserID, 10)

			// Пропускаем плательщика
			if userTransaction.UserID == transaction.PayerID {
				continue
			}

			// Конвертируем сумму из float64 в int (умножаем на 100 для сохранения копеек)
			amountInt := int(userTransaction.UserPart * 100)

			// Добавляем долг в карту
			if _, exists := debts[payerKey][debtorKey]; !exists {
				debts[payerKey][debtorKey] = 0
			}
			debts[payerKey][debtorKey] += amountInt
		}
	}

	// Оптимизируем долги
	optimizedDebts := SimplifyDebts(debts)

	// Преобразуем оптимизированные долги в ответ API
	var result []dto.EventTransactionTemporalResponse
	var totalID int = 1

	for creditorKey, debtors := range optimizedDebts {
		for debtorKey, amount := range debtors {
			creditorID, _ := strconv.ParseInt(creditorKey, 10, 64)
			creditor, exists := userMap[creditorID]
			if !exists {
				continue
			}

			debtorID, _ := strconv.ParseInt(debtorKey, 10, 64)
			debtor, debtorExists := userMap[debtorID]
			if !debtorExists {
				continue
			}

			responseItem := dto.EventTransactionTemporalResponse{
				TotalID: totalID,
				Amount:  float64(amount) / 100, // Конвертируем обратно в float64
			}

			// Заполняем данные кредитора
			responseItem.Requestor.ID = creditor.IDUser
			responseItem.Requestor.Name = creditor.NameCashed
			responseItem.Requestor.Photo = creditor.PhotoUUIDCashed

			// Заполняем данные должника
			responseItem.Debtor.ID = debtor.IDUser
			responseItem.Debtor.Name = debtor.NameCashed
			responseItem.Debtor.Photo = debtor.PhotoUUIDCashed

			result = append(result, responseItem)
			totalID++
		}
	}

	return result, nil
}

// mapTransactionToResponse преобразует модель Transaction в DTO
func mapTransactionToResponse(transaction models.Transaction) dto.TransactionResponse {
	return dto.TransactionResponse{
		TransactionID: transaction.ID,
		Name:          transaction.Name,
		TransactionTypeID: struct {
			ID     int    `json:"id"`
			IconID string `json:"icon_id"`
		}{
			ID:     transaction.TransactionTypeID,
			IconID: "", // Заполняется в сервисе при необходимости
		},
		DateTime:  transaction.DateTime,
		TotalPaid: transaction.TotalPaid,
		PayerName: transaction.Payer.NameCashed,
	}
}

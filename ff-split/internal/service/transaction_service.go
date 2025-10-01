package service

import (
	"context"
	"time"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres"
	"github.com/ivasnev/FinFlow/ff-split/internal/service/debt_calculator"
	"gorm.io/gorm"
)

// TransactionService реализует сервис для работы с транзакциями
type TransactionService struct {
	db           *gorm.DB
	repo         *postgres.TransactionRepository
	userService  UserServiceInterface
	eventService EventServiceInterface
}

// NewTransactionService создает новый сервис для работы с транзакциями
func NewTransactionService(
	db *gorm.DB,
	repo *postgres.TransactionRepository,
	userService UserServiceInterface,
	eventService EventServiceInterface,
) *TransactionService {
	return &TransactionService{
		db:           db,
		repo:         repo,
		userService:  userService,
		eventService: eventService,
	}
}

// GetTransactionsByEventID возвращает список транзакций мероприятия
func (s *TransactionService) GetTransactionsByEventID(ctx context.Context, eventID int64) ([]dto.TransactionResponse, error) {
	// Проверяем существование мероприятия
	_, err := s.eventService.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Получаем транзакции
	transactions, err := s.repo.GetTransactionsByEventID(eventID)
	if err != nil {
		return nil, err
	}

	// Преобразуем в DTO
	result := make([]dto.TransactionResponse, len(transactions))
	for i, tx := range transactions {
		// Получаем доли
		shares, err := s.repo.GetSharesByTransactionID(tx.ID)
		if err != nil {
			return nil, err
		}

		// Получаем долги
		debts, err := s.repo.GetDebtsByTransactionID(tx.ID)
		if err != nil {
			return nil, err
		}

		// Преобразуем в DTO
		txResponse, err := s.mapTransactionToDTO(&tx, shares, debts)
		if err != nil {
			return nil, err
		}

		result[i] = *txResponse
	}

	return result, nil
}

// GetTransactionByID возвращает транзакцию по ID
func (s *TransactionService) GetTransactionByID(ctx context.Context, id int) (*dto.TransactionResponse, error) {
	// Получаем транзакцию
	tx, err := s.repo.GetTransactionByID(id)
	if err != nil {
		return nil, err
	}

	// Получаем доли
	shares, err := s.repo.GetSharesByTransactionID(tx.ID)
	if err != nil {
		return nil, err
	}

	// Получаем долги
	debts, err := s.repo.GetDebtsByTransactionID(tx.ID)
	if err != nil {
		return nil, err
	}

	// Преобразуем в DTO
	return s.mapTransactionToDTO(tx, shares, debts)
}

// CreateTransaction создает новую транзакцию
func (s *TransactionService) CreateTransaction(ctx context.Context, eventID int64, req *dto.TransactionRequest) (*dto.TransactionResponse, error) {
	// Начинаем транзакцию в базе данных
	var result *dto.TransactionResponse
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Проверяем существование мероприятия
		_, err := s.eventService.GetEventByID(ctx, eventID)
		if err != nil {
			return err
		}

		// Проверяем, что плательщик существует
		payer, err := s.userService.GetUserByInternalUserID(ctx, req.FromUser)
		if err != nil {
			return err
		}

		// Создаем запись о транзакции
		transaction := &models.Transaction{
			EventID:               &eventID,
			Name:                  req.Name,
			TransactionCategoryID: req.TransactionCategoryID,
			Datetime:              time.Now(),
			TotalPaid:             req.Amount,
			PayerID:               &payer.ID,
			SplitType:             s.getSplitTypeID(req.Type),
		}

		if err := s.repo.CreateTransaction(transaction); err != nil {
			return err
		}

		// Рассчитываем доли и долги
		calculator, err := debt_calculator.GetCalculator(req.Type)
		if err != nil {
			return err
		}

		shares, debts, err := calculator.Calculate(req, eventID)
		if err != nil {
			return err
		}

		// Преобразуем внутренние Share в модель TransactionShare
		dbShares := make([]models.TransactionShare, len(shares))
		for i, share := range shares {
			dbShares[i] = models.TransactionShare{
				TransactionID: transaction.ID,
				UserID:        share.UserID,
				Value:         share.Value,
			}
		}

		// Сохраняем доли в базе
		if err := s.repo.CreateTransactionShares(dbShares); err != nil {
			return err
		}

		// Преобразуем внутренние Debt в модель Debt
		dbDebts := make([]models.Debt, len(debts))
		for i, debt := range debts {
			dbDebts[i] = models.Debt{
				TransactionID: transaction.ID,
				FromUserID:    debt.FromUserID,
				ToUserID:      debt.ToUserID,
				Amount:        debt.Amount,
			}
		}

		// Сохраняем долги в базе
		if err := s.repo.CreateDebts(dbDebts); err != nil {
			return err
		}

		// Формируем ответ
		resp, err := s.mapTransactionToDTO(transaction, dbShares, dbDebts)
		if err != nil {
			return err
		}

		result = resp
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateTransaction обновляет существующую транзакцию
func (s *TransactionService) UpdateTransaction(ctx context.Context, id int, req *dto.TransactionRequest) (*dto.TransactionResponse, error) {
	// Начинаем транзакцию в базе данных
	var result *dto.TransactionResponse
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Получаем транзакцию
		transaction, err := s.repo.GetTransactionByID(id)
		if err != nil {
			return err
		}

		// Проверяем, что плательщик существует
		payer, err := s.userService.GetUserByInternalUserID(ctx, req.FromUser)
		if err != nil {
			return err
		}

		// Обновляем данные транзакции
		transaction.Name = req.Name
		transaction.TransactionCategoryID = req.TransactionCategoryID
		transaction.TotalPaid = req.Amount
		transaction.PayerID = &payer.ID
		transaction.SplitType = s.getSplitTypeID(req.Type)

		if err := s.repo.UpdateTransaction(transaction); err != nil {
			return err
		}

		// Удаляем старые доли и долги
		if err := s.repo.DeleteSharesByTransactionID(id); err != nil {
			return err
		}

		if err := s.repo.DeleteDebtsByTransactionID(id); err != nil {
			return err
		}

		// Рассчитываем новые доли и долги
		calculator, err := debt_calculator.GetCalculator(req.Type)
		if err != nil {
			return err
		}

		eventID := *transaction.EventID
		shares, debts, err := calculator.Calculate(req, eventID)
		if err != nil {
			return err
		}

		// Преобразуем внутренние Share в модель TransactionShare
		dbShares := make([]models.TransactionShare, len(shares))
		for i, share := range shares {
			dbShares[i] = models.TransactionShare{
				TransactionID: transaction.ID,
				UserID:        share.UserID,
				Value:         share.Value,
			}
		}

		// Сохраняем доли в базе
		if err := s.repo.CreateTransactionShares(dbShares); err != nil {
			return err
		}

		// Преобразуем внутренние Debt в модель Debt
		dbDebts := make([]models.Debt, len(debts))
		for i, debt := range debts {
			dbDebts[i] = models.Debt{
				TransactionID: transaction.ID,
				FromUserID:    debt.FromUserID,
				ToUserID:      debt.ToUserID,
				Amount:        debt.Amount,
			}
		}

		// Сохраняем долги в базе
		if err := s.repo.CreateDebts(dbDebts); err != nil {
			return err
		}

		// Формируем ответ
		resp, err := s.mapTransactionToDTO(transaction, dbShares, dbDebts)
		if err != nil {
			return err
		}

		result = resp
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteTransaction удаляет транзакцию
func (s *TransactionService) DeleteTransaction(ctx context.Context, id int) error {
	return s.repo.DeleteTransaction(id)
}

// GetDebtsByEventID возвращает долги в рамках мероприятия
func (s *TransactionService) GetDebtsByEventID(ctx context.Context, eventID int64) ([]dto.DebtDTO, error) {
	// Проверяем существование мероприятия
	_, err := s.eventService.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Получаем долги
	debts, err := s.repo.GetDebtsByEventID(eventID)
	if err != nil {
		return nil, err
	}

	// Преобразуем в DTO
	result := make([]dto.DebtDTO, len(debts))
	for i, debt := range debts {
		result[i] = dto.DebtDTO{
			ID:            debt.ID,
			FromUserID:    debt.FromUserID,
			ToUserID:      debt.ToUserID,
			Amount:        debt.Amount,
			TransactionID: debt.TransactionID,
		}
	}

	return result, nil
}

// Вспомогательные методы

// mapTransactionToDTO преобразует модель Transaction в DTO
func (s *TransactionService) mapTransactionToDTO(
	tx *models.Transaction,
	shares []models.TransactionShare,
	debts []models.Debt,
) (*dto.TransactionResponse, error) {
	// Преобразуем доли в DTO
	shareDTOs := make([]dto.ShareDTO, len(shares))
	for i, share := range shares {
		shareDTOs[i] = dto.ShareDTO{
			ID:            share.ID,
			UserID:        share.UserID,
			Value:         share.Value,
			TransactionID: share.TransactionID,
		}
	}

	// Преобразуем долги в DTO
	debtDTOs := make([]dto.DebtDTO, len(debts))
	for i, debt := range debts {
		debtDTOs[i] = dto.DebtDTO{
			ID:            debt.ID,
			FromUserID:    debt.FromUserID,
			ToUserID:      debt.ToUserID,
			Amount:        debt.Amount,
			TransactionID: debt.TransactionID,
		}
	}

	// Формируем ответ
	var eventID int64
	if tx.EventID != nil {
		eventID = *tx.EventID
	}

	var fromUser int64
	if tx.PayerID != nil {
		user, err := s.userService.GetUserByInternalUserID(context.Background(), *tx.PayerID)
		if err != nil {
			return nil, err
		}
		if user.UserID != nil {
			fromUser = *user.UserID
		}
	}

	return &dto.TransactionResponse{
		ID:                    tx.ID,
		EventID:               eventID,
		Name:                  tx.Name,
		TransactionCategoryID: tx.TransactionCategoryID,
		Type:                  s.getSplitTypeName(tx.SplitType),
		FromUser:              fromUser,
		Amount:                tx.TotalPaid,
		Datetime:              tx.Datetime,
		Debts:                 debtDTOs,
		Shares:                shareDTOs,
	}, nil
}

// getSplitTypeID преобразует строковое представление типа распределения в числовой ID
func (s *TransactionService) getSplitTypeID(splitType string) int {
	switch splitType {
	case debt_calculator.PercentType:
		return 1
	case debt_calculator.AmountType:
		return 2
	case debt_calculator.UnitsType:
		return 3
	default:
		return 0 // По умолчанию (поровну)
	}
}

// getSplitTypeName преобразует числовой ID типа распределения в строковое представление
func (s *TransactionService) getSplitTypeName(splitTypeID int) string {
	switch splitTypeID {
	case 1:
		return debt_calculator.PercentType
	case 2:
		return debt_calculator.AmountType
	case 3:
		return debt_calculator.UnitsType
	default:
		return "equal" // По умолчанию (поровну)
	}
}

package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres"
	"github.com/ivasnev/FinFlow/ff-split/internal/service/debs_optimizer"
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
func (s *TransactionService) GetDebtsByEventID(ctx context.Context, eventID int64, userID *int64) ([]dto.DebtDTO, error) {
	// Проверяем существование мероприятия
	_, err := s.eventService.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	var debts []dto.DebtDTO
	if userID == nil {
		// Получаем долги
		eventDebts, err := s.repo.GetDebtsByEventID(eventID)
		if err != nil {
			return nil, err
		}
		// Преобразуем в DTO
		for _, debt := range eventDebts {
			debtDTO := dto.DebtDTO{
				ID:            debt.ID,
				FromUserID:    debt.FromUserID,
				ToUserID:      debt.ToUserID,
				Amount:        debt.Amount,
				TransactionID: debt.TransactionID,
			}
			if debt.FromUser != nil {
				debtDTO.FromUser = &dto.DebtsUserResponse{
					ID:         debt.FromUser.ID,
					ExternalID: debt.FromUser.UserID,
					Name:       debt.FromUser.NameCashed,
					Photo:      debt.FromUser.PhotoUUIDCashed,
				}
			}
			if debt.ToUser != nil {
				debtDTO.ToUser = &dto.DebtsUserResponse{
					ID:         debt.ToUser.ID,
					ExternalID: debt.ToUser.UserID,
					Name:       debt.ToUser.NameCashed,
					Photo:      debt.ToUser.PhotoUUIDCashed,
				}
			}
			debts = append(debts, debtDTO)
		}
	} else {
		user, err := s.userService.GetUserByExternalUserID(ctx, *userID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
		}
		debtsToUser, err := s.GetDebtsByEventIDToUser(eventID, user.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении долгов пользователю: %w", err)
		}
		debts = append(debts, debtsToUser...)
		debtsFromUser, err := s.GetDebtsByEventIDFromUser(eventID, user.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении долгов пользователю: %w", err)
		}
		debts = append(debts, debtsFromUser...)
	}

	return debts, nil
}

func (s *TransactionService) GetDebtsByEventIDFromUser(eventID int64, userID int64) ([]dto.DebtDTO, error) {
	debtsFromUser, err := s.repo.GetDebtsByEventIDFromUser(eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении долгов пользователю: %w", err)
	}
	var result []dto.DebtDTO
	for _, debt := range debtsFromUser {
		result = append(result, dto.DebtDTO{
			ID:            debt.ID,
			FromUserID:    debt.FromUserID,
			ToUserID:      debt.ToUserID,
			Amount:        -debt.Amount,
			TransactionID: debt.TransactionID,

			Requestor: &dto.DebtsUserResponse{
				ID:         debt.ToUser.ID,
				ExternalID: debt.ToUser.UserID,
				Name:       debt.ToUser.NameCashed,
				Photo:      debt.ToUser.PhotoUUIDCashed,
			},
		})
	}
	return result, nil
}

func (s *TransactionService) GetDebtsByEventIDToUser(eventID int64, userID int64) ([]dto.DebtDTO, error) {
	debtsToUser, err := s.repo.GetDebtsByEventIDToUser(eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении долгов пользователю: %w", err)
	}
	var result []dto.DebtDTO
	for _, debt := range debtsToUser {
		result = append(result, dto.DebtDTO{
			ID:            debt.ID,
			FromUserID:    debt.FromUserID,
			ToUserID:      debt.ToUserID,
			Amount:        debt.Amount,
			TransactionID: debt.TransactionID,

			Requestor: &dto.DebtsUserResponse{
				ID:         debt.ToUser.ID,
				ExternalID: debt.ToUser.UserID,
				Name:       debt.ToUser.NameCashed,
				Photo:      debt.ToUser.PhotoUUIDCashed,
			},
		})
	}
	return result, nil
}

// OptimizeDebts оптимизирует долги для мероприятия и сохраняет результат
func (s *TransactionService) OptimizeDebts(ctx context.Context, eventID int64) ([]dto.OptimizedDebtDTO, error) {
	// Проверяем существование мероприятия
	_, err := s.eventService.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Получаем все долги мероприятия
	debts, err := s.repo.GetDebtsByEventID(eventID)
	if err != nil {
		return nil, err
	}

	// Формируем структуру для оптимизатора
	debtMap := make(map[string]map[string]int)
	for _, debt := range debts {
		fromUserID := strconv.FormatInt(debt.FromUserID, 10)
		toUserID := strconv.FormatInt(debt.ToUserID, 10)

		if _, exists := debtMap[toUserID]; !exists {
			debtMap[toUserID] = make(map[string]int)
		}

		// Округляем до целых для алгоритма оптимизации
		amount := int(math.Round(debt.Amount))
		debtMap[toUserID][fromUserID] += amount
	}

	// Оптимизируем долги
	optimizedDebts := debs_optimizer.SimplifyDebts(debtMap)

	// Преобразуем результат в модель и DTO
	result := make([]dto.OptimizedDebtDTO, 0)
	modelsToSave := make([]models.OptimizedDebt, 0)

	for creditor, debtors := range optimizedDebts {
		for debtor, amount := range debtors {
			if amount <= 0 {
				continue
			}

			creditorID, _ := strconv.ParseInt(creditor, 10, 64)
			debtorID, _ := strconv.ParseInt(debtor, 10, 64)

			// Создаем модель для сохранения
			optimizedDebt := models.OptimizedDebt{
				EventID:    eventID,
				FromUserID: debtorID,
				ToUserID:   creditorID,
				Amount:     float64(amount),
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}
			modelsToSave = append(modelsToSave, optimizedDebt)

			// Создаем DTO для ответа
			result = append(result, dto.OptimizedDebtDTO{
				FromUserID: debtorID,
				ToUserID:   creditorID,
				Amount:     float64(amount),
				EventID:    eventID,
			})
		}
	}

	// Сохраняем оптимизированные долги в базе
	if err := s.repo.SaveOptimizedDebts(eventID, modelsToSave); err != nil {
		return nil, err
	}

	return result, nil
}

// GetOptimizedDebtsByEventID возвращает оптимизированные долги по ID мероприятия
func (s *TransactionService) GetOptimizedDebtsByEventID(ctx context.Context, eventID int64) ([]dto.OptimizedDebtDTO, error) {
	// Проверяем существование мероприятия
	_, err := s.eventService.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Получаем оптимизированные долги
	optimizedDebts, err := s.repo.GetOptimizedDebtsByEventID(eventID)
	if err != nil {
		return nil, err
	}

	// Если оптимизированных долгов нет, вызываем оптимизацию
	if len(optimizedDebts) == 0 {
		return s.OptimizeDebts(ctx, eventID)
	}

	// Формируем ответ
	result := make([]dto.OptimizedDebtDTO, len(optimizedDebts))
	for i, debt := range optimizedDebts {
		result[i] = dto.OptimizedDebtDTO{
			ID:         debt.ID,
			FromUserID: debt.FromUserID,
			ToUserID:   debt.ToUserID,
			Amount:     debt.Amount,
			EventID:    debt.EventID,
		}
	}

	return result, nil
}

// GetOptimizedDebtsByUserID возвращает оптимизированные долги по ID пользователя в мероприятии
func (s *TransactionService) GetOptimizedDebtsByUserID(ctx context.Context, eventID, userID int64) ([]dto.OptimizedDebtDTO, error) {
	// Проверяем существование мероприятия
	_, err := s.eventService.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Проверяем существование пользователя
	_, err = s.userService.GetUserByInternalUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Получаем оптимизированные долги
	optimizedDebts, err := s.repo.GetOptimizedDebtsByUserID(eventID, userID)
	if err != nil {
		return nil, err
	}

	// Если оптимизированных долгов нет, вызываем оптимизацию и затем фильтруем
	if len(optimizedDebts) == 0 {
		allDebts, err := s.OptimizeDebts(ctx, eventID)
		if err != nil {
			return nil, err
		}

		// Фильтруем долги, связанные с пользователем
		userDebts := make([]dto.OptimizedDebtDTO, 0)
		for _, debt := range allDebts {
			if debt.FromUserID == userID || debt.ToUserID == userID {
				userDebts = append(userDebts, debt)
			}
		}
		return userDebts, nil
	}

	// Формируем ответ
	result := make([]dto.OptimizedDebtDTO, len(optimizedDebts))
	for i, debt := range optimizedDebts {
		result[i] = dto.OptimizedDebtDTO{
			ID:         debt.ID,
			FromUserID: debt.FromUserID,
			ToUserID:   debt.ToUserID,
			Amount:     debt.Amount,
			EventID:    debt.EventID,
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

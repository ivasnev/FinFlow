package transaction

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/ivasnev/FinFlow/ff-common/optimizers"
	"github.com/ivasnev/FinFlow/ff-common/optimizers/dinic"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	"github.com/ivasnev/FinFlow/ff-split/internal/service/debt_calculator"
	"gorm.io/gorm"
)

// TransactionService реализует сервис для работы с транзакциями
type TransactionService struct {
	db           *gorm.DB
	repo         repository.Transaction
	userService  service.User
	eventService service.Event
}

// NewTransactionService создает новый сервис для работы с транзакциями
func NewTransactionService(
	db *gorm.DB,
	repo repository.Transaction,
	userService service.User,
	eventService service.Event,
) *TransactionService {
	return &TransactionService{
		db:           db,
		repo:         repo,
		userService:  userService,
		eventService: eventService,
	}
}

// GetTransactionsByEventID возвращает список транзакций мероприятия
func (s *TransactionService) GetTransactionsByEventID(ctx context.Context, eventID int64) ([]service.TransactionResponse, error) {
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
	result := make([]service.TransactionResponse, len(transactions))
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
func (s *TransactionService) GetTransactionByID(ctx context.Context, id int) (*service.TransactionResponse, error) {
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
func (s *TransactionService) CreateTransaction(ctx context.Context, eventID int64, req *service.TransactionRequest) (*service.TransactionResponse, error) {
	// Начинаем транзакцию в базе данных
	var result *service.TransactionResponse
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
func (s *TransactionService) UpdateTransaction(ctx context.Context, id int, req *service.TransactionRequest) (*service.TransactionResponse, error) {
	// Начинаем транзакцию в базе данных
	var result *service.TransactionResponse
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
func (s *TransactionService) GetDebtsByEventID(ctx context.Context, eventID int64, userID *int64) ([]service.DebtDTO, error) {
	// Проверяем существование мероприятия
	_, err := s.eventService.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}
	var debts []service.DebtDTO
	if userID == nil {
		// Получаем долги
		eventDebts, err := s.repo.GetDebtsByEventID(eventID)
		if err != nil {
			return nil, err
		}
		// Преобразуем в DTO
		for _, debt := range eventDebts {
			debtDTO := service.DebtDTO{
				ID:            debt.ID,
				FromUserID:    debt.FromUserID,
				ToUserID:      debt.ToUserID,
				Amount:        debt.Amount,
				TransactionID: debt.TransactionID,
			}
			if debt.FromUser != nil {
				debtDTO.FromUser = &service.DebtsUserResponse{
					ID:    debt.FromUser.ID,
					Name:  getUserName(debt.FromUser),
					Photo: debt.FromUser.PhotoUUIDCashed,
				}
			}
			if debt.ToUser != nil {
				debtDTO.ToUser = &service.DebtsUserResponse{
					ID:    debt.ToUser.ID,
					Name:  getUserName(debt.ToUser),
					Photo: debt.ToUser.PhotoUUIDCashed,
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

func (s *TransactionService) GetDebtsByEventIDFromUser(eventID int64, userID int64) ([]service.DebtDTO, error) {
	debtsFromUser, err := s.repo.GetDebtsByEventIDFromUser(eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении долгов пользователю: %w", err)
	}
	var result []service.DebtDTO
	for _, debt := range debtsFromUser {
		result = append(result, service.DebtDTO{
			ID:            debt.ID,
			FromUserID:    debt.FromUserID,
			ToUserID:      debt.ToUserID,
			Amount:        -debt.Amount,
			TransactionID: debt.TransactionID,

			Requestor: &service.DebtsUserResponse{
				ID:    debt.ToUser.ID,
				Name:  getUserName(debt.ToUser),
				Photo: debt.ToUser.PhotoUUIDCashed,
			},
		})
	}
	return result, nil
}

func (s *TransactionService) GetDebtsByEventIDToUser(eventID int64, userID int64) ([]service.DebtDTO, error) {
	debtsToUser, err := s.repo.GetDebtsByEventIDToUser(eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении долгов пользователю: %w", err)
	}
	var result []service.DebtDTO
	for _, debt := range debtsToUser {
		result = append(result, service.DebtDTO{
			ID:            debt.ID,
			FromUserID:    debt.FromUserID,
			ToUserID:      debt.ToUserID,
			Amount:        debt.Amount,
			TransactionID: debt.TransactionID,

			Requestor: &service.DebtsUserResponse{
				ID:    debt.FromUser.ID,
				Name:  getUserName(debt.FromUser),
				Photo: debt.FromUser.PhotoUUIDCashed,
			},
		})
	}
	return result, nil
}

func getUserName(user *models.User) string {
	if user.NameCashed != "" {
		return user.NameCashed
	} else if user.NicknameCashed != "" {
		return user.NicknameCashed
	} else {
		return "Incognito"
	}
}

// OptimizeDebts оптимизирует долги для мероприятия и сохраняет результат
func (s *TransactionService) OptimizeDebts(ctx context.Context, eventID int64) ([]service.OptimizedDebtDTO, error) {
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

	// Собираем переводы для оптимизатора (From — должник, To — кредитор)
	transfers := make([]optimizers.Transfer, 0, len(debts))
	for _, debt := range debts {
		transfers = append(transfers, optimizers.Transfer{
			From:   strconv.FormatInt(debt.FromUserID, 10),
			To:     strconv.FormatInt(debt.ToUserID, 10),
			Amount: int(math.Round(debt.Amount)),
		})
	}

	optimized, err := dinic.New().Optimize(transfers)
	if err != nil {
		return nil, err
	}

	result := make([]service.OptimizedDebtDTO, 0, len(optimized))
	modelsToSave := make([]models.OptimizedDebt, 0, len(optimized))

	for _, t := range optimized {
		if t.Amount <= 0 {
			continue
		}

		fromID, _ := strconv.ParseInt(t.From, 10, 64)
		toID, _ := strconv.ParseInt(t.To, 10, 64)

		modelsToSave = append(modelsToSave, models.OptimizedDebt{
			EventID:    eventID,
			FromUserID: fromID,
			ToUserID:   toID,
			Amount:     float64(t.Amount),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		})
		result = append(result, service.OptimizedDebtDTO{
			FromUserID: fromID,
			ToUserID:   toID,
			Amount:     float64(t.Amount),
			EventID:    eventID,
		})
	}

	// Сохраняем оптимизированные долги в базе
	if err := s.repo.SaveOptimizedDebts(eventID, modelsToSave); err != nil {
		return nil, err
	}

	return result, nil
}

// GetOptimizedDebtsByEventID возвращает оптимизированные долги по ID мероприятия
func (s *TransactionService) GetOptimizedDebtsByEventID(ctx context.Context, eventID int64, userID *int64) ([]service.OptimizedDebtDTO, error) {
	// Проверяем существование мероприятия
	_, err := s.eventService.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	var debts []service.OptimizedDebtDTO
	if userID == nil {
		// Получаем все оптимизированные долги мероприятия
		optimizedDebts, err := s.repo.GetOptimizedDebtsByEventIDWithUsers(eventID)
		if err != nil {
			return nil, err
		}

		// Если оптимизированных долгов нет, вызываем оптимизацию
		if len(optimizedDebts) == 0 {
			return s.OptimizeDebts(ctx, eventID)
		}

		// Преобразуем в DTO
		for _, debt := range optimizedDebts {
			debtDTO := service.OptimizedDebtDTO{
				ID:         debt.ID,
				FromUserID: debt.FromUserID,
				ToUserID:   debt.ToUserID,
				Amount:     debt.Amount,
				EventID:    debt.EventID,
			}
			if debt.FromUser != nil {
				debtDTO.FromUser = &service.DebtsUserResponse{
					ID:    debt.FromUser.ID,
					Name:  getUserName(debt.FromUser),
					Photo: debt.FromUser.PhotoUUIDCashed,
				}
			}
			if debt.ToUser != nil {
				debtDTO.ToUser = &service.DebtsUserResponse{
					ID:    debt.ToUser.ID,
					Name:  getUserName(debt.ToUser),
					Photo: debt.ToUser.PhotoUUIDCashed,
				}
			}
			debts = append(debts, debtDTO)
		}
	} else {
		user, err := s.userService.GetUserByExternalUserID(ctx, *userID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
		}
		debtsToUser, err := s.GetOptimizedDebtsByEventIDToUser(eventID, user.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении оптимизированных долгов пользователю: %w", err)
		}
		debts = append(debts, debtsToUser...)
		debtsFromUser, err := s.GetOptimizedDebtsByEventIDFromUser(eventID, user.ID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при получении оптимизированных долгов от пользователя: %w", err)
		}
		debts = append(debts, debtsFromUser...)
	}

	return debts, nil
}

// GetOptimizedDebtsByEventIDFromUser возвращает оптимизированные долги от пользователя в мероприятии
func (s *TransactionService) GetOptimizedDebtsByEventIDFromUser(eventID int64, userID int64) ([]service.OptimizedDebtDTO, error) {
	optimizedDebts, err := s.repo.GetOptimizedDebtsByUserIDWithUsers(eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении оптимизированных долгов от пользователя: %w", err)
	}

	var result []service.OptimizedDebtDTO
	for _, debt := range optimizedDebts {
		if debt.FromUserID == userID {
			result = append(result, service.OptimizedDebtDTO{
				ID:         debt.ID,
				FromUserID: debt.FromUserID,
				ToUserID:   debt.ToUserID,
				Amount:     -debt.Amount,
				EventID:    debt.EventID,

				Requestor: &service.DebtsUserResponse{
					ID:    debt.ToUser.ID,
					Name:  getUserName(debt.ToUser),
					Photo: debt.ToUser.PhotoUUIDCashed,
				},
			})
		}
	}
	return result, nil
}

// GetOptimizedDebtsByEventIDToUser возвращает оптимизированные долги к пользователю в мероприятии
func (s *TransactionService) GetOptimizedDebtsByEventIDToUser(eventID int64, userID int64) ([]service.OptimizedDebtDTO, error) {
	optimizedDebts, err := s.repo.GetOptimizedDebtsByUserIDWithUsers(eventID, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении оптимизированных долгов к пользователю: %w", err)
	}

	var result []service.OptimizedDebtDTO
	for _, debt := range optimizedDebts {
		if debt.ToUserID == userID {
			result = append(result, service.OptimizedDebtDTO{
				ID:         debt.ID,
				FromUserID: debt.FromUserID,
				ToUserID:   debt.ToUserID,
				Amount:     debt.Amount,
				EventID:    debt.EventID,

				Requestor: &service.DebtsUserResponse{
					ID:    debt.FromUser.ID,
					Name:  getUserName(debt.FromUser),
					Photo: debt.FromUser.PhotoUUIDCashed,
				},
			})
		}
	}
	return result, nil
}

// GetOptimizedDebtsByUserID возвращает оптимизированные долги по ID пользователя в мероприятии
func (s *TransactionService) GetOptimizedDebtsByUserID(ctx context.Context, eventID, userID int64) ([]service.OptimizedDebtDTO, error) {
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
	optimizedDebts, err := s.repo.GetOptimizedDebtsByUserIDWithUsers(eventID, userID)
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
		userDebts := make([]service.OptimizedDebtDTO, 0)
		for _, debt := range allDebts {
			if debt.FromUserID == userID || debt.ToUserID == userID {
				userDebts = append(userDebts, debt)
			}
		}
		return userDebts, nil
	}

	// Формируем ответ
	result := make([]service.OptimizedDebtDTO, len(optimizedDebts))
	for i, debt := range optimizedDebts {
		debtDTO := service.OptimizedDebtDTO{
			ID:         debt.ID,
			FromUserID: debt.FromUserID,
			ToUserID:   debt.ToUserID,
			Amount:     debt.Amount,
			EventID:    debt.EventID,
		}
		if debt.FromUser != nil {
			debtDTO.FromUser = &service.DebtsUserResponse{
				ID:    debt.FromUser.ID,
				Name:  getUserName(debt.FromUser),
				Photo: debt.FromUser.PhotoUUIDCashed,
			}
		}
		if debt.ToUser != nil {
			debtDTO.ToUser = &service.DebtsUserResponse{
				ID:    debt.ToUser.ID,
				Name:  getUserName(debt.ToUser),
				Photo: debt.ToUser.PhotoUUIDCashed,
			}
		}
		result[i] = debtDTO
	}

	return result, nil
}

// Вспомогательные методы

// mapTransactionToDTO преобразует модель Transaction в DTO
func (s *TransactionService) mapTransactionToDTO(
	tx *models.Transaction,
	shares []models.TransactionShare,
	debts []models.Debt,
) (*service.TransactionResponse, error) {
	// Преобразуем доли в DTO
	shareDTOs := make([]service.ShareDTO, len(shares))
	for i, share := range shares {
		shareDTOs[i] = service.ShareDTO{
			ID:            share.ID,
			UserID:        share.UserID,
			Value:         share.Value,
			TransactionID: share.TransactionID,
		}
	}

	// Преобразуем долги в DTO
	debtDTOs := make([]service.DebtDTO, len(debts))
	for i, debt := range debts {
		debtDTOs[i] = service.DebtDTO{
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
		// Всегда используем внутренний ID
		fromUser = *tx.PayerID
	}

	return &service.TransactionResponse{
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

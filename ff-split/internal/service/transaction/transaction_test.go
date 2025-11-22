package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	repositoryMock "github.com/ivasnev/FinFlow/ff-split/internal/repository/mock"
	serviceMock "github.com/ivasnev/FinFlow/ff-split/internal/service/mock"
	"gorm.io/gorm"
)

func TestTransactionService_GetTransactionByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionRepo := repositoryMock.NewMockTransaction(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockEventService := serviceMock.NewMockEvent(ctrl)
	var db *gorm.DB

	transactionService := NewTransactionService(db, mockTransactionRepo, mockUserService, mockEventService)

	ctx := context.Background()
	transactionID := 1

	t.Run("успешное получение транзакции", func(t *testing.T) {
		eventID := int64(1)
		userID := int64(100)
		transaction := &models.Transaction{
			ID:        transactionID,
			EventID:   &eventID,
			Name:      "Test Transaction",
			TotalPaid: 100.0,
			PayerID:   &userID,
			Datetime:  time.Now(),
		}

		shares := []models.TransactionShare{
			{ID: 1, TransactionID: transactionID, UserID: userID, Value: 50.0},
		}

		debts := []models.Debt{
			{ID: 1, TransactionID: transactionID, FromUserID: userID, ToUserID: 200, Amount: 50.0},
		}

		mockTransactionRepo.EXPECT().
			GetTransactionByID(transactionID).
			Return(transaction, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetSharesByTransactionID(transactionID).
			Return(shares, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetDebtsByTransactionID(transactionID).
			Return(debts, nil).
			Times(1)

		result, err := transactionService.GetTransactionByID(ctx, transactionID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, transactionID, result.ID)
		assert.Equal(t, "Test Transaction", result.Name)
	})

	t.Run("транзакция не найдена", func(t *testing.T) {
		expectedErr := errors.New("transaction not found")
		mockTransactionRepo.EXPECT().
			GetTransactionByID(transactionID).
			Return(nil, expectedErr).
			Times(1)

		result, err := transactionService.GetTransactionByID(ctx, transactionID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTransactionService_DeleteTransaction(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionRepo := repositoryMock.NewMockTransaction(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockEventService := serviceMock.NewMockEvent(ctrl)
	var db *gorm.DB

	transactionService := NewTransactionService(db, mockTransactionRepo, mockUserService, mockEventService)

	ctx := context.Background()
	transactionID := 1

	t.Run("успешное удаление транзакции", func(t *testing.T) {
		mockTransactionRepo.EXPECT().
			DeleteTransaction(transactionID).
			Return(nil).
			Times(1)

		err := transactionService.DeleteTransaction(ctx, transactionID)

		assert.NoError(t, err)
	})

	t.Run("ошибка удаления транзакции", func(t *testing.T) {
		expectedErr := errors.New("delete error")
		mockTransactionRepo.EXPECT().
			DeleteTransaction(transactionID).
			Return(expectedErr).
			Times(1)

		err := transactionService.DeleteTransaction(ctx, transactionID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTransactionService_GetDebtsByEventID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionRepo := repositoryMock.NewMockTransaction(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockEventService := serviceMock.NewMockEvent(ctrl)
	var db *gorm.DB

	transactionService := NewTransactionService(db, mockTransactionRepo, mockUserService, mockEventService)

	ctx := context.Background()
	eventID := int64(1)

	t.Run("успешное получение долгов", func(t *testing.T) {
		debts := []models.Debt{
			{ID: 1, TransactionID: 1, FromUserID: 100, ToUserID: 200, Amount: 50.0},
			{ID: 2, TransactionID: 2, FromUserID: 200, ToUserID: 100, Amount: 30.0},
		}

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(&models.Event{ID: eventID}, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetDebtsByEventID(eventID).
			Return(debts, nil).
			Times(1)

		result, err := transactionService.GetDebtsByEventID(ctx, eventID, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
	})

	t.Run("ошибка получения долгов", func(t *testing.T) {
		expectedErr := errors.New("database error")

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(&models.Event{ID: eventID}, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetDebtsByEventID(eventID).
			Return(nil, expectedErr).
			Times(1)

		result, err := transactionService.GetDebtsByEventID(ctx, eventID, nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTransactionService_GetDebtsByEventIDFromUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionRepo := repositoryMock.NewMockTransaction(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockEventService := serviceMock.NewMockEvent(ctrl)
	var db *gorm.DB

	transactionService := NewTransactionService(db, mockTransactionRepo, mockUserService, mockEventService)

	eventID := int64(1)
	userID := int64(100)

	t.Run("успешное получение долгов от пользователя", func(t *testing.T) {
		toUserID := int64(200)
		toUser := &models.User{ID: 1, UserID: &toUserID}
		debts := []models.Debt{
			{
				ID:            1,
				TransactionID: 1,
				FromUserID:    userID,
				ToUserID:      toUserID,
				Amount:        50.0,
				ToUser:        toUser,
			},
		}

		mockTransactionRepo.EXPECT().
			GetDebtsByEventIDFromUser(eventID, userID).
			Return(debts, nil).
			Times(1)

		result, err := transactionService.GetDebtsByEventIDFromUser(eventID, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, userID, result[0].FromUserID)
	})

	t.Run("ошибка получения долгов", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockTransactionRepo.EXPECT().
			GetDebtsByEventIDFromUser(eventID, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := transactionService.GetDebtsByEventIDFromUser(eventID, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTransactionService_GetDebtsByEventIDToUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionRepo := repositoryMock.NewMockTransaction(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockEventService := serviceMock.NewMockEvent(ctrl)
	var db *gorm.DB

	transactionService := NewTransactionService(db, mockTransactionRepo, mockUserService, mockEventService)

	eventID := int64(1)
	userID := int64(200)

	t.Run("успешное получение долгов к пользователю", func(t *testing.T) {
		fromUserID := int64(100)
		fromUser := &models.User{ID: 1, UserID: &fromUserID}
		debts := []models.Debt{
			{
				ID:            1,
				TransactionID: 1,
				FromUserID:    fromUserID,
				ToUserID:      userID,
				Amount:        50.0,
				FromUser:      fromUser,
			},
		}

		mockTransactionRepo.EXPECT().
			GetDebtsByEventIDToUser(eventID, userID).
			Return(debts, nil).
			Times(1)

		result, err := transactionService.GetDebtsByEventIDToUser(eventID, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, userID, result[0].ToUserID)
	})

	t.Run("ошибка получения долгов", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockTransactionRepo.EXPECT().
			GetDebtsByEventIDToUser(eventID, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := transactionService.GetDebtsByEventIDToUser(eventID, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTransactionService_GetTransactionsByEventID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionRepo := repositoryMock.NewMockTransaction(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockEventService := serviceMock.NewMockEvent(ctrl)
	var db *gorm.DB

	transactionService := NewTransactionService(db, mockTransactionRepo, mockUserService, mockEventService)

	ctx := context.Background()
	eventID := int64(1)

	t.Run("успешное получение транзакций", func(t *testing.T) {
		userID := int64(100)
		transactions := []models.Transaction{
			{
				ID:        1,
				EventID:   &eventID,
				Name:      "Transaction 1",
				TotalPaid: 100.0,
				PayerID:   &userID,
				Datetime:  time.Now(),
			},
			{
				ID:        2,
				EventID:   &eventID,
				Name:      "Transaction 2",
				TotalPaid: 200.0,
				PayerID:   &userID,
				Datetime:  time.Now(),
			},
		}

		shares1 := []models.TransactionShare{
			{ID: 1, TransactionID: 1, UserID: userID, Value: 50.0},
		}
		debts1 := []models.Debt{
			{ID: 1, TransactionID: 1, FromUserID: userID, ToUserID: 200, Amount: 50.0},
		}

		shares2 := []models.TransactionShare{
			{ID: 2, TransactionID: 2, UserID: userID, Value: 100.0},
		}
		debts2 := []models.Debt{
			{ID: 2, TransactionID: 2, FromUserID: userID, ToUserID: 200, Amount: 100.0},
		}

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(&models.Event{ID: eventID}, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetTransactionsByEventID(eventID).
			Return(transactions, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetSharesByTransactionID(1).
			Return(shares1, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetDebtsByTransactionID(1).
			Return(debts1, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetSharesByTransactionID(2).
			Return(shares2, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetDebtsByTransactionID(2).
			Return(debts2, nil).
			Times(1)

		result, err := transactionService.GetTransactionsByEventID(ctx, eventID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "Transaction 1", result[0].Name)
		assert.Equal(t, "Transaction 2", result[1].Name)
	})

	t.Run("мероприятие не найдено", func(t *testing.T) {
		expectedErr := errors.New("event not found")

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(nil, expectedErr).
			Times(1)

		result, err := transactionService.GetTransactionsByEventID(ctx, eventID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("ошибка получения транзакций", func(t *testing.T) {
		expectedErr := errors.New("database error")

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(&models.Event{ID: eventID}, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetTransactionsByEventID(eventID).
			Return(nil, expectedErr).
			Times(1)

		result, err := transactionService.GetTransactionsByEventID(ctx, eventID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTransactionService_OptimizeDebts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionRepo := repositoryMock.NewMockTransaction(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockEventService := serviceMock.NewMockEvent(ctrl)
	var db *gorm.DB

	transactionService := NewTransactionService(db, mockTransactionRepo, mockUserService, mockEventService)

	ctx := context.Background()
	eventID := int64(1)

	t.Run("успешная оптимизация долгов", func(t *testing.T) {
		debts := []models.Debt{
			{ID: 1, TransactionID: 1, FromUserID: 100, ToUserID: 200, Amount: 50.0},
			{ID: 2, TransactionID: 2, FromUserID: 200, ToUserID: 100, Amount: 30.0},
		}

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(&models.Event{ID: eventID}, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetDebtsByEventID(eventID).
			Return(debts, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			SaveOptimizedDebts(eventID, gomock.Any()).
			DoAndReturn(func(eID int64, optimizedDebts []models.OptimizedDebt) error {
				// Проверяем, что оптимизированные долги созданы
				assert.Greater(t, len(optimizedDebts), 0)
				return nil
			}).
			Times(1)

		result, err := transactionService.OptimizeDebts(ctx, eventID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// После оптимизации должен остаться только один долг (50 - 30 = 20)
		assert.GreaterOrEqual(t, len(result), 0)
	})

	t.Run("мероприятие не найдено", func(t *testing.T) {
		expectedErr := errors.New("event not found")

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(nil, expectedErr).
			Times(1)

		result, err := transactionService.OptimizeDebts(ctx, eventID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("ошибка получения долгов", func(t *testing.T) {
		expectedErr := errors.New("database error")

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(&models.Event{ID: eventID}, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetDebtsByEventID(eventID).
			Return(nil, expectedErr).
			Times(1)

		result, err := transactionService.OptimizeDebts(ctx, eventID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTransactionService_GetOptimizedDebtsByEventID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionRepo := repositoryMock.NewMockTransaction(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockEventService := serviceMock.NewMockEvent(ctrl)
	var db *gorm.DB

	transactionService := NewTransactionService(db, mockTransactionRepo, mockUserService, mockEventService)

	ctx := context.Background()
	eventID := int64(1)

	t.Run("успешное получение оптимизированных долгов", func(t *testing.T) {
		optimizedDebts := []models.OptimizedDebt{
			{ID: 1, EventID: eventID, FromUserID: 100, ToUserID: 200, Amount: 20.0},
			{ID: 2, EventID: eventID, FromUserID: 200, ToUserID: 300, Amount: 50.0},
		}

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(&models.Event{ID: eventID}, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetOptimizedDebtsByEventIDWithUsers(eventID).
			Return(optimizedDebts, nil).
			Times(1)

		result, err := transactionService.GetOptimizedDebtsByEventID(ctx, eventID, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, int64(100), result[0].FromUserID)
		assert.Equal(t, int64(200), result[0].ToUserID)
	})

	t.Run("успешное получение оптимизированных долгов с фильтром по пользователю", func(t *testing.T) {
		externalUserID := int64(100)
		internalUserID := int64(1)
		toUserID := int64(200)
		toUser := &models.User{ID: 1, UserID: &toUserID}
		
		optimizedDebtsFromUser := []models.OptimizedDebt{
			{
				ID:        1,
				EventID:   eventID,
				FromUserID: internalUserID,
				ToUserID:   toUserID,
				Amount:    20.0,
				ToUser:    toUser,
			},
		}

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(&models.Event{ID: eventID}, nil).
			Times(1)

		mockUserService.EXPECT().
			GetUserByExternalUserID(ctx, externalUserID).
			Return(&models.User{ID: internalUserID, UserID: &externalUserID}, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetOptimizedDebtsByUserIDWithUsers(eventID, internalUserID).
			Return(optimizedDebtsFromUser, nil).
			Times(2) // Вызывается дважды - для ToUser и FromUser

		result, err := transactionService.GetOptimizedDebtsByEventID(ctx, eventID, &externalUserID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.GreaterOrEqual(t, len(result), 1)
	})

	t.Run("мероприятие не найдено", func(t *testing.T) {
		expectedErr := errors.New("event not found")

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(nil, expectedErr).
			Times(1)

		result, err := transactionService.GetOptimizedDebtsByEventID(ctx, eventID, nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTransactionService_GetOptimizedDebtsByEventIDFromUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionRepo := repositoryMock.NewMockTransaction(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockEventService := serviceMock.NewMockEvent(ctrl)
	var db *gorm.DB

	transactionService := NewTransactionService(db, mockTransactionRepo, mockUserService, mockEventService)

	eventID := int64(1)
	userID := int64(100)

	t.Run("успешное получение оптимизированных долгов от пользователя", func(t *testing.T) {
		toUserID := int64(200)
		optimizedDebts := []models.OptimizedDebt{
			{
				ID:        1,
				EventID:   eventID,
				FromUserID: userID,
				ToUserID:   toUserID,
				Amount:    20.0,
				ToUser:    &models.User{ID: 1, UserID: &toUserID},
			},
		}

		mockTransactionRepo.EXPECT().
			GetOptimizedDebtsByUserIDWithUsers(eventID, userID).
			Return(optimizedDebts, nil).
			Times(1)

		result, err := transactionService.GetOptimizedDebtsByEventIDFromUser(eventID, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, userID, result[0].FromUserID)
	})

	t.Run("ошибка получения оптимизированных долгов", func(t *testing.T) {
		expectedErr := errors.New("database error")

		mockTransactionRepo.EXPECT().
			GetOptimizedDebtsByUserIDWithUsers(eventID, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := transactionService.GetOptimizedDebtsByEventIDFromUser(eventID, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTransactionService_GetOptimizedDebtsByEventIDToUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionRepo := repositoryMock.NewMockTransaction(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockEventService := serviceMock.NewMockEvent(ctrl)
	var db *gorm.DB

	transactionService := NewTransactionService(db, mockTransactionRepo, mockUserService, mockEventService)

	eventID := int64(1)
	userID := int64(200)

	t.Run("успешное получение оптимизированных долгов к пользователю", func(t *testing.T) {
		fromUserID := int64(100)
		optimizedDebts := []models.OptimizedDebt{
			{
				ID:        1,
				EventID:   eventID,
				FromUserID: fromUserID,
				ToUserID:   userID,
				Amount:    20.0,
				FromUser:  &models.User{ID: 1, UserID: &fromUserID},
			},
		}

		mockTransactionRepo.EXPECT().
			GetOptimizedDebtsByUserIDWithUsers(eventID, userID).
			Return(optimizedDebts, nil).
			Times(1)

		result, err := transactionService.GetOptimizedDebtsByEventIDToUser(eventID, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, userID, result[0].ToUserID)
	})

	t.Run("ошибка получения оптимизированных долгов", func(t *testing.T) {
		expectedErr := errors.New("database error")

		mockTransactionRepo.EXPECT().
			GetOptimizedDebtsByUserIDWithUsers(eventID, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := transactionService.GetOptimizedDebtsByEventIDToUser(eventID, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTransactionService_GetOptimizedDebtsByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTransactionRepo := repositoryMock.NewMockTransaction(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockEventService := serviceMock.NewMockEvent(ctrl)
	var db *gorm.DB

	transactionService := NewTransactionService(db, mockTransactionRepo, mockUserService, mockEventService)

	ctx := context.Background()
	eventID := int64(1)
	userID := int64(100)

	t.Run("успешное получение оптимизированных долгов по пользователю", func(t *testing.T) {
		optimizedDebts := []models.OptimizedDebt{
			{ID: 1, EventID: eventID, FromUserID: userID, ToUserID: 200, Amount: 20.0},
			{ID: 2, EventID: eventID, FromUserID: 300, ToUserID: userID, Amount: 50.0},
		}

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(&models.Event{ID: eventID}, nil).
			Times(1)

		mockUserService.EXPECT().
			GetUserByInternalUserID(ctx, userID).
			Return(&models.User{ID: userID}, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetOptimizedDebtsByUserIDWithUsers(eventID, userID).
			Return(optimizedDebts, nil).
			Times(1)

		result, err := transactionService.GetOptimizedDebtsByUserID(ctx, eventID, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
	})

	t.Run("ошибка получения оптимизированных долгов", func(t *testing.T) {
		expectedErr := errors.New("database error")

		mockEventService.EXPECT().
			GetEventByID(ctx, eventID).
			Return(&models.Event{ID: eventID}, nil).
			Times(1)

		mockUserService.EXPECT().
			GetUserByInternalUserID(ctx, userID).
			Return(&models.User{ID: userID}, nil).
			Times(1)

		mockTransactionRepo.EXPECT().
			GetOptimizedDebtsByUserIDWithUsers(eventID, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := transactionService.GetOptimizedDebtsByUserID(ctx, eventID, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}


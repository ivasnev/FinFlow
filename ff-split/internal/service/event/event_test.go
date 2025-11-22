package event

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	repositoryMock "github.com/ivasnev/FinFlow/ff-split/internal/repository/mock"
	serviceMock "github.com/ivasnev/FinFlow/ff-split/internal/service/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestEventService_GetEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventRepo := repositoryMock.NewMockEvent(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockCategoryService := serviceMock.NewMockCategory(ctrl)
	var db *gorm.DB // В реальных тестах можно использовать тестовую БД

	eventService := NewEventService(mockEventRepo, db, mockUserService, mockCategoryService)

	ctx := context.Background()

	t.Run("успешное получение мероприятий", func(t *testing.T) {
		events := []models.Event{
			{ID: 1, Name: "Event 1", Status: "active"},
			{ID: 2, Name: "Event 2", Status: "active"},
		}

		mockEventRepo.EXPECT().
			GetAll(ctx).
			Return(events, nil).
			Times(1)

		result, err := eventService.GetEvents(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
	})

	t.Run("ошибка получения мероприятий", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockEventRepo.EXPECT().
			GetAll(ctx).
			Return(nil, expectedErr).
			Times(1)

		result, err := eventService.GetEvents(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestEventService_GetEventByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventRepo := repositoryMock.NewMockEvent(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockCategoryService := serviceMock.NewMockCategory(ctrl)
	var db *gorm.DB

	eventService := NewEventService(mockEventRepo, db, mockUserService, mockCategoryService)

	ctx := context.Background()
	eventID := int64(1)

	t.Run("успешное получение мероприятия", func(t *testing.T) {
		event := &models.Event{
			ID:     eventID,
			Name:   "Test Event",
			Status: "active",
		}

		mockEventRepo.EXPECT().
			GetByID(ctx, eventID).
			Return(event, nil).
			Times(1)

		result, err := eventService.GetEventByID(ctx, eventID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, eventID, result.ID)
		assert.Equal(t, "Test Event", result.Name)
	})

	t.Run("мероприятие не найдено", func(t *testing.T) {
		expectedErr := errors.New("event not found")
		mockEventRepo.EXPECT().
			GetByID(ctx, eventID).
			Return(nil, expectedErr).
			Times(1)

		result, err := eventService.GetEventByID(ctx, eventID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestEventService_GetBalanceByEventID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEventRepo := repositoryMock.NewMockEvent(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockCategoryService := serviceMock.NewMockCategory(ctrl)
	var db *gorm.DB

	eventService := NewEventService(mockEventRepo, db, mockUserService, mockCategoryService)

	ctx := context.Background()
	userID := int64(100)
	eventID := int64(1)

	t.Run("успешное получение баланса", func(t *testing.T) {
		balances := map[int64]float64{
			eventID: 150.50,
		}

		mockEventRepo.EXPECT().
			CalculateUserBalances(ctx, userID, []int64{eventID}).
			Return(balances, nil).
			Times(1)

		result, err := eventService.GetBalanceByEventID(ctx, userID, eventID)

		assert.NoError(t, err)
		assert.Equal(t, 150.50, result)
	})

	t.Run("баланс не найден", func(t *testing.T) {
		balances := map[int64]float64{}

		mockEventRepo.EXPECT().
			CalculateUserBalances(ctx, userID, []int64{eventID}).
			Return(balances, nil).
			Times(1)

		result, err := eventService.GetBalanceByEventID(ctx, userID, eventID)

		assert.NoError(t, err)
		assert.Equal(t, 0.0, result)
	})

	t.Run("ошибка расчета баланса", func(t *testing.T) {
		expectedErr := errors.New("calculation error")
		mockEventRepo.EXPECT().
			CalculateUserBalances(ctx, userID, []int64{eventID}).
			Return(nil, expectedErr).
			Times(1)

		result, err := eventService.GetBalanceByEventID(ctx, userID, eventID)

		assert.Error(t, err)
		assert.Equal(t, 0.0, result)
		assert.Contains(t, err.Error(), "ошибка при расчете баланса")
	})
}

func TestEventService_DeleteEvent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Создаем in-memory SQLite БД для тестов с транзакциями
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Ошибка создания тестовой БД: %v", err)
	}

	mockEventRepo := repositoryMock.NewMockEvent(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	mockCategoryService := serviceMock.NewMockCategory(ctrl)

	eventService := NewEventService(mockEventRepo, testDB, mockUserService, mockCategoryService)

	ctx := context.Background()
	eventID := int64(1)

	t.Run("успешное удаление мероприятия", func(t *testing.T) {
		mockEventRepo.EXPECT().
			Delete(gomock.Any(), eventID).
			Return(nil).
			Times(1)

		err := eventService.DeleteEvent(ctx, eventID)

		assert.NoError(t, err)
	})

	t.Run("ошибка удаления мероприятия", func(t *testing.T) {
		expectedErr := errors.New("delete error")
		mockEventRepo.EXPECT().
			Delete(gomock.Any(), eventID).
			Return(expectedErr).
			Times(1)

		err := eventService.DeleteEvent(ctx, eventID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Ошибка при удалении мероприятия")
	})
}


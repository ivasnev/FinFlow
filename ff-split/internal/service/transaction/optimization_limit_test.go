package transaction

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	domainerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	repositoryMock "github.com/ivasnev/FinFlow/ff-split/internal/repository/mock"
	serviceMock "github.com/ivasnev/FinFlow/ff-split/internal/service/mock"
)

// TestOptimizationLimit проверяет отказ без очереди при занятых вычислительных слотах.
func TestOptimizationLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	repo := repositoryMock.NewMockTransaction(ctrl)
	events := serviceMock.NewMockEvent(ctrl)
	events.EXPECT().GetEventByID(gomock.Any(), int64(1)).Return(&models.Event{ID: 1}, nil)
	repo.EXPECT().WithContext(gomock.Any()).Return(repo).AnyTimes()
	repo.EXPECT().GetOptimizationState(int64(1)).Return(&models.DebtOptimizationState{}, nil).AnyTimes()
	for i := 0; i < cap(optimizationSlots); i++ {
		optimizationSlots <- struct{}{}
	}
	defer func() {
		for i := 0; i < cap(optimizationSlots); i++ {
			<-optimizationSlots
		}
	}()
	svc := NewTransactionService(nil, repo, nil, events)
	if _, err := svc.OptimizeDebts(context.Background(), 1); !errors.Is(err, domainerrors.ErrOptimizationBusy) {
		t.Fatalf("overload was not rejected: %v", err)
	}
}

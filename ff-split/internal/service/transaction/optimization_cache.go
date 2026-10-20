package transaction

import (
	"context"
	"strconv"
	"time"

	domainerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// optimizationSlots ограничивает дорогие расчёты во всём процессе; очереди ожидания слота нет.
var optimizationSlots = make(chan struct{}, 2)

// optimizedModels объединяет одновременные запросы и кэширует в БД даже пустой результат.
// Общим расчётом владеет сервис: отмена отдельного ожидания не отменяет работу остальных.
// Модели остаются приватными и используются только для чтения; каждый ответ получает новые DTO.
func (s *TransactionService) optimizedModels(ctx context.Context, eventID int64) ([]models.OptimizedDebt, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// Версия читается каждым запросом: читатель после commit не присоединяется к старому снимку.
	callerRepo := s.repo.WithContext(ctx)
	observed, err := callerRepo.GetOptimizationState(eventID)
	if err != nil {
		return nil, err
	}
	if observed.OptimizedDebtRevision != nil && *observed.OptimizedDebtRevision == observed.DebtRevision {
		return callerRepo.GetOptimizedDebtsByEventIDWithUsers(eventID)
	}
	key := strconv.FormatInt(eventID, 10) + ":" + strconv.FormatInt(observed.DebtRevision, 10)
	result := s.optimization.DoChan(key, func() (interface{}, error) {
		workCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()
		repo := s.repo.WithContext(workCtx)
		state, err := repo.GetOptimizationState(eventID)
		if err != nil {
			return nil, err
		}
		if state.OptimizedDebtRevision != nil && *state.OptimizedDebtRevision == state.DebtRevision {
			return repo.GetOptimizedDebtsByEventIDWithUsers(eventID)
		}
		select {
		case optimizationSlots <- struct{}{}:
			defer func() { <-optimizationSlots }()
		default:
			return nil, domainerrors.ErrOptimizationBusy
		}
		var debts []models.OptimizedDebt
		err = repo.RunOptimization(workCtx, eventID, func(bound repository.Transaction) error {
			// Другой процесс мог завершить расчёт между первым чтением и получением замка.
			state, err := bound.GetOptimizationState(eventID)
			if err != nil {
				return err
			}
			if state.OptimizedDebtRevision != nil && *state.OptimizedDebtRevision == state.DebtRevision {
				debts, err = bound.GetOptimizedDebtsByEventIDWithUsers(eventID)
			} else {
				debts, err = s.computeOptimizedDebts(workCtx, bound, eventID, state.DebtRevision)
			}
			return err
		})
		return debts, err
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case shared := <-result:
		if shared.Err != nil {
			return nil, shared.Err
		}
		return shared.Val.([]models.OptimizedDebt), nil
	}
}

// mapOptimizedDebts формирует независимые ответы из приватного результата общего расчёта.
func mapOptimizedDebts(debts []models.OptimizedDebt) []service.OptimizedDebtDTO {
	result := make([]service.OptimizedDebtDTO, 0, len(debts))
	for _, debt := range debts {
		dto := service.OptimizedDebtDTO{ID: debt.ID, EventID: debt.EventID, FromUserID: debt.FromUserID, ToUserID: debt.ToUserID, Amount: debt.Amount}
		if debt.FromUser != nil {
			dto.FromUser = &service.DebtsUserResponse{ID: debt.FromUser.ID, Name: getUserName(debt.FromUser), Photo: debt.FromUser.PhotoUUIDCashed}
		}
		if debt.ToUser != nil {
			dto.ToUser = &service.DebtsUserResponse{ID: debt.ToUser.ID, Name: getUserName(debt.ToUser), Photo: debt.ToUser.PhotoUUIDCashed}
		}
		result = append(result, dto)
	}
	return result
}

package settlement

import (
	"context"
	"math"

	"github.com/google/uuid"
	domainerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// Service учитывает только подтверждённое получателем фактическое поступление денег.
type Service struct{ repo repository.Settlement }

// NewService создаёт сервис подтверждённых возвратов.
func NewService(repo repository.Settlement) *Service { return &Service{repo: repo} }

// List возвращает историю возвратов доступного мероприятия.
func (s *Service) List(ctx context.Context, eventID, before int64) ([]models.Settlement, error) {
	return s.repo.List(ctx, eventID, before)
}

// Receive сериализует лимит остатка, ключ повтора и сохранение записи.
func (s *Service) Receive(ctx context.Context, eventID, externalActor int64, req service.SettlementRequest) (*models.Settlement, error) {
	key, err := uuid.Parse(req.RequestID)
	if err != nil || key == uuid.Nil || req.FromUserID <= 0 || req.AmountCents <= 0 || req.AmountCents > 9999999999 {
		return nil, domainerrors.NewValidationError("settlement", "проверьте участника, положительную сумму до копейки и ключ запроса")
	}
	var result *models.Settlement
	err = s.repo.InEvent(ctx, eventID, func(ctx context.Context) error {
		members, err := s.repo.Members(ctx, eventID)
		if err != nil {
			return err
		}
		var actorID int64
		for id, external := range members {
			if external != nil && *external == externalActor {
				actorID = id
				break
			}
		}
		if actorID == 0 {
			return domainerrors.ErrForbidden
		}
		if req.FromUserID == actorID {
			return domainerrors.NewValidationError("from_user_id", "получатель и отправитель должны различаться")
		}
		previous, err := s.repo.ByRequest(ctx, eventID, actorID, key.String())
		if err != nil {
			return err
		}
		if previous != nil {
			if previous.FromUserID != req.FromUserID || previous.AmountCents != req.AmountCents {
				return domainerrors.NewAlreadyExistsError("settlement", "этот ключ уже использован для другого возврата")
			}
			result = previous
			return nil
		}
		if _, exists := members[req.FromUserID]; !exists {
			return domainerrors.NewValidationError("from_user_id", "отправитель должен участвовать в событии")
		}
		ledger, err := s.repo.Ledger(ctx, eventID)
		if err != nil {
			return err
		}
		balances := map[int64]int64{}
		var pair int64
		for _, debt := range ledger {
			cents := int64(math.Round(debt.Amount * 100))
			balances[debt.FromUserID] -= cents
			balances[debt.ToUserID] += cents
			if debt.FromUserID == req.FromUserID && debt.ToUserID == actorID {
				pair += cents
			}
			if debt.FromUserID == actorID && debt.ToUserID == req.FromUserID {
				pair -= cents
			}
		}
		net := min(max(-balances[req.FromUserID], int64(0)), max(balances[actorID], int64(0)))
		if req.AmountCents > max(pair, net) {
			return domainerrors.NewValidationError("amount_cents", "сумма возврата превышает текущий долг; обновите расчёты")
		}
		result = &models.Settlement{EventID: eventID, FromUserID: req.FromUserID, ToUserID: actorID, CreatedBy: actorID, AmountCents: req.AmountCents, RequestID: key.String()}
		return s.repo.Create(ctx, result)
	})
	return result, err
}

// Cancel отменяет ошибку создателя без повторной проверки уже изменившегося денежного лимита.
func (s *Service) Cancel(ctx context.Context, eventID, externalActor, id int64) error {
	return s.repo.InEvent(ctx, eventID, func(ctx context.Context) error {
		row, err := s.repo.Get(ctx, eventID, id)
		if err != nil {
			return err
		}
		members, err := s.repo.Members(ctx, eventID)
		if err != nil {
			return err
		}
		external := members[row.CreatedBy]
		if external == nil || *external != externalActor {
			return domainerrors.ErrForbidden
		}
		return s.repo.Cancel(ctx, eventID, id)
	})
}

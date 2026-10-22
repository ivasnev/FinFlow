package receipt

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	"time"

	commonerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Retry повторяет только завершённое распознавание и не заменяет сохранённые правки.
func (s *Service) Retry(ctx context.Context, eventID int64, id string, revision int) (*models.Receipt, error) {
	if revision < 1 {
		return nil, commonerrors.NewValidationError("receipt", "укажите ревизию чека")
	}
	before, err := s.Get(ctx, eventID, id)
	if err != nil {
		return nil, err
	}
	if before.RetryRevision != revision {
		if !before.ImageAvailable {
			return nil, commonerrors.NewValidationError("receipt", "фотография больше недоступна; загрузите её заново")
		}
		if before.Revision != revision || (before.Status != "ready" && before.Status != "failed") {
			return nil, commonerrors.NewValidationError("receipt", "чек уже изменён или ещё распознаётся; обновите его")
		}
		release, ok := s.ImageSlot()
		if !ok {
			return nil, commonerrors.NewValidationError("receipt", "загрузка занята; повторите позже")
		}
		defer release()
		actorID, _ := actor.ID(ctx)
		if err := s.repo.AdmitImageRead(ctx, actorID); err != nil {
			return nil, err
		}
		read, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		image, err := s.store.Read(read, before.ImageKey)
		if err != nil || len(image) == 0 {
			return nil, commonerrors.NewValidationError("receipt", "фотография недоступна; загрузите её заново")
		}
	}
	var result *models.Receipt
	err = s.repo.InEvent(ctx, eventID, func(ctx context.Context) error {
		actorID, err := s.member(ctx, eventID)
		if err != nil {
			return err
		}
		v, err := s.repo.Get(ctx, eventID, id, true)
		if err != nil {
			return err
		}
		// Повтор того же HTTP-запроса не расходует квоту даже после завершения OCR.
		if v.RetryRevision == revision {
			result = v
			return nil
		}
		if v.Revision != revision || (v.Status != "ready" && v.Status != "failed") {
			return commonerrors.NewValidationError("receipt", "чек уже изменён или ещё распознаётся; обновите его")
		}
		if err := s.repo.Retry(ctx, v, actorID); err != nil {
			return err
		}
		result = v
		return nil
	})
	return result, err
}

// Remove запрещает удалять чек, по которому уже созданы расходы.
func (s *Service) Remove(ctx context.Context, eventID int64, id string, revision int) error {
	if revision < 1 {
		return commonerrors.NewValidationError("receipt", "укажите ревизию чека")
	}
	return s.repo.InEvent(ctx, eventID, func(ctx context.Context) error {
		if _, err := s.member(ctx, eventID); err != nil {
			return err
		}
		v, err := s.repo.Get(ctx, eventID, id, true)
		if err != nil {
			return err
		}
		if v.Status == "confirmed" {
			return commonerrors.NewValidationError("receipt", "расходы по этому чеку уже созданы")
		}
		if v.Revision != revision {
			return commonerrors.NewValidationError("receipt", "чек уже изменён; обновите его")
		}
		return s.repo.Remove(ctx, v)
	})
}

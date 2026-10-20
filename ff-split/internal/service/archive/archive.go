package archive

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
)

// Service разрешает участникам восстановить событие до ночной окончательной очистки.
type Service struct{ repo repository.Archive }

// NewService создаёт сервис обратимого архивирования.
func NewService(repo repository.Archive) *Service { return &Service{repo: repo} }

// List отдаёт только архив событий текущего участника.
func (s *Service) List(ctx context.Context) ([]models.Event, error) {
	id, ok := actor.ID(ctx)
	if !ok {
		return nil, errors.ErrForbidden
	}
	return s.repo.List(ctx, id)
}

// change проверяет участие внутри блокировки и не меняет срок при повторном запросе.
func (s *Service) change(ctx context.Context, id int64, archived bool) error {
	actor, ok := actor.ID(ctx)
	if !ok {
		return errors.ErrForbidden
	}
	return s.repo.InEvent(ctx, id, func(ctx context.Context, row *models.Event) error {
		member, err := s.repo.IsMember(ctx, id, actor)
		if err != nil {
			return err
		}
		if !member {
			return errors.ErrForbidden
		}
		if (row.DeletedAt != nil) == archived {
			return nil
		}
		return s.repo.SetArchived(ctx, row, archived)
	})
}

// Archive помечает событие вместо физического удаления.
func (s *Service) Archive(ctx context.Context, id int64) error { return s.change(ctx, id, true) }

// Restore возвращает сохранённые расходы, задачи и участников без воссоздания данных.
func (s *Service) Restore(ctx context.Context, id int64) error { return s.change(ctx, id, false) }

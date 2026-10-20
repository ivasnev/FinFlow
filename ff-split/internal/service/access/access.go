package access

import (
	"context"
	"errors"
)

// ErrDenied обозначает отсутствие доступа к мероприятию или его объекту.
var ErrDenied = errors.New("event access denied")

// Repository предоставляет факты об участии и принадлежности объектов.
type Repository interface {
	// IsMember проверяет участие по внешнему ID пользователя.
	IsMember(ctx context.Context, externalID, eventID int64) (bool, error)
	// HasResource проверяет привязку вложенного объекта.
	HasResource(ctx context.Context, eventID int64, kind string, resourceID int64) (bool, error)
}

// Service разрешает операции участникам только внутри их мероприятий.
type Service struct{ repo Repository }

// NewService создаёт сервис проверки доступа.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Authorize проверяет актёра и, при наличии, вложенный объект.
func (s *Service) Authorize(ctx context.Context, externalID, eventID int64, kind string, resourceID int64) error {
	member, err := s.repo.IsMember(ctx, externalID, eventID)
	if err != nil {
		return err
	}
	if !member {
		return ErrDenied
	}
	if kind != "" {
		belongs, err := s.repo.HasResource(ctx, eventID, kind, resourceID)
		if err != nil {
			return err
		}
		if !belongs {
			return ErrDenied
		}
	}
	return nil
}

package invitation

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/ivasnev/FinFlow/ff-split/internal/common/actor"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// Service создаёт ограниченные ссылки и добавляет только подтвердившего вход пользователя.
type Service struct {
	repo  repository.Invitation
	users service.User
}

// NewService подключает хранилище и синхронизацию профиля из ID-сервиса.
func NewService(repo repository.Invitation, users service.User) *Service {
	return &Service{repo: repo, users: users}
}

// hashToken проверяет формат и не сохраняет исходный секрет.
func hashToken(token string) (string, error) {
	if len(token) != 43 {
		return "", errors.ErrInvitationUnavailable
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 32 {
		return "", errors.ErrInvitationUnavailable
	}
	hash := sha256.Sum256(raw)
	return hex.EncodeToString(hash[:]), nil
}

// member не доверяет проверке до блокировки события при изменении его состава.
func (s *Service) member(ctx context.Context, eventID int64) (int64, error) {
	id, ok := actor.ID(ctx)
	if !ok {
		return 0, errors.ErrForbidden
	}
	member, err := s.repo.IsMember(ctx, eventID, id)
	if err != nil {
		return 0, err
	}
	if !member {
		return 0, errors.ErrForbidden
	}
	return id, nil
}

// List не раскрывает ссылки другого создателя.
func (s *Service) List(ctx context.Context, eventID int64) ([]models.Invitation, error) {
	id, err := s.member(ctx, eventID)
	if err != nil {
		return nil, err
	}
	return s.repo.List(ctx, eventID, id)
}

// Create возвращает исходный секрет ровно один раз и ограничивает число активных ссылок.
func (s *Service) Create(ctx context.Context, eventID int64) (*models.Invitation, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	hash, _ := hashToken(token)
	var row *models.Invitation
	err := s.repo.InEvent(ctx, eventID, func(ctx context.Context) error {
		id, err := s.member(ctx, eventID)
		if err != nil {
			return err
		}
		active, err := s.repo.List(ctx, eventID, id)
		if err != nil {
			return err
		}
		if len(active) >= 5 {
			return errors.NewValidationError("invitation", "у вас уже пять действующих ссылок; отзовите ненужную")
		}
		row = &models.Invitation{EventID: eventID, CreatedBy: id, TokenHash: hash}
		return s.repo.Create(ctx, row)
	})
	if err != nil {
		return nil, "", err
	}
	return row, token, nil
}

// Revoke разрешает отзыв только создателю ссылки и только в исходном событии.
func (s *Service) Revoke(ctx context.Context, eventID, id int64) error {
	return s.repo.InEvent(ctx, eventID, func(ctx context.Context) error {
		actor, err := s.member(ctx, eventID)
		if err != nil {
			return err
		}
		row, err := s.repo.Get(ctx, eventID, id)
		if err != nil {
			return err
		}
		if row.CreatedBy != actor {
			return errors.ErrForbidden
		}
		return s.repo.Revoke(ctx, row)
	})
}

// Preview не добавляет участника при открытии страницы или предпросмотре ссылки.
func (s *Service) Preview(ctx context.Context, token string) (*models.InvitationPreview, error) {
	id, ok := actor.ID(ctx)
	if !ok {
		return nil, errors.ErrForbidden
	}
	hash, err := hashToken(token)
	if err != nil {
		return nil, err
	}
	row, err := s.repo.Lookup(ctx, hash)
	if err != nil {
		return nil, err
	}
	name, err := s.repo.EventName(ctx, row.EventID)
	if err != nil {
		return nil, err
	}
	member, err := s.repo.IsMember(ctx, row.EventID, id)
	if err != nil {
		return nil, err
	}
	return &models.InvitationPreview{EventID: row.EventID, EventName: name, ExpiresAt: row.ExpiresAt, IsMember: member}, nil
}

// Accept повторно проверяет срок и отзыв после блокировки, затем атомарно записывает участие.
func (s *Service) Accept(ctx context.Context, token string) (int64, error) {
	id, ok := actor.ID(ctx)
	if !ok {
		return 0, errors.ErrForbidden
	}
	hash, err := hashToken(token)
	if err != nil {
		return 0, err
	}
	row, err := s.repo.Lookup(ctx, hash)
	if err != nil {
		return 0, err
	}
	user, err := s.users.GetUserByExternalUserID(ctx, id)
	if err != nil {
		return 0, err
	}
	if user.IsDummy {
		return 0, errors.ErrForbidden
	}
	err = s.repo.InEvent(ctx, row.EventID, func(ctx context.Context) error {
		current, err := s.repo.Lookup(ctx, hash)
		if err != nil {
			return err
		}
		return s.repo.Join(ctx, current.EventID, user.ID)
	})
	if err != nil {
		return 0, err
	}
	return row.EventID, nil
}

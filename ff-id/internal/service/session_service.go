package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/dto"
	"github.com/ivasnev/FinFlow/ff-id/interfaces"
)

// SessionService реализует интерфейс для работы с сессиями
type SessionService struct {
	sessionRepository interfaces.SessionRepository
}

// NewSessionService создает новый сервис сессий
func NewSessionService(
	sessionRepository interfaces.SessionRepository,
) *SessionService {
	return &SessionService{
		sessionRepository: sessionRepository,
	}
}

// GetUserSessions получает все сессии пользователя
func (s *SessionService) GetUserSessions(ctx context.Context, userID int64) ([]dto.SessionDTO, error) {
	sessions, err := s.sessionRepository.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения сессий: %w", err)
	}

	// Преобразуем в DTO
	result := make([]dto.SessionDTO, len(sessions))
	for i, session := range sessions {
		result[i] = dto.SessionDTO{
			ID:        session.ID,
			IPAddress: string(session.IPAddress[0]), // Берем первый IP для отображения
			CreatedAt: session.CreatedAt,
			ExpiresAt: session.ExpiresAt,
		}
	}

	return result, nil
}

// TerminateSession завершает сессию
func (s *SessionService) TerminateSession(ctx context.Context, sessionID uuid.UUID, userID int64) error {
	// Получаем сессию по ID
	session, err := s.sessionRepository.GetByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("ошибка получения сессии: %w", err)
	}

	// Проверяем, принадлежит ли сессия пользователю
	if session.UserID != userID {
		return errors.New("у вас нет прав на удаление этой сессии")
	}

	// Удаляем сессию
	return s.sessionRepository.Delete(ctx, sessionID)
}

// TerminateAllSessions завершает все сессии пользователя
func (s *SessionService) TerminateAllSessions(ctx context.Context, userID int64) error {
	return s.sessionRepository.DeleteAllByUserID(ctx, userID)
}

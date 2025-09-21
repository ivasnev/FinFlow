package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/ivasnev/FinFlow/ff-id/internal/repository/postgres"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/api/dto"
)

// SessionService реализует интерфейс для работы с сессиями
type SessionService struct {
	sessionRepository postgres.SessionRepositoryInterface
}

// NewSessionService создает новый сервис сессий
func NewSessionService(
	sessionRepository postgres.SessionRepositoryInterface,
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
		var ipAdress string
		if session.IPAddress != nil {
			ipAdress = session.IPAddress[0]
		}
		result[i] = dto.SessionDTO{
			ID:        session.ID,
			IPAddress: ipAdress, // Берем первый IP для отображения
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

package service

import (
	"context"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-id/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-id/internal/repository/postgres"
)

// LoginHistoryService реализует интерфейс для работы с историей входов
type LoginHistoryService struct {
	loginHistoryRepository postgres.LoginHistoryRepositoryInterface
}

// NewLoginHistoryService создает новый сервис истории входов
func NewLoginHistoryService(
	loginHistoryRepository postgres.LoginHistoryRepositoryInterface,
) *LoginHistoryService {
	return &LoginHistoryService{
		loginHistoryRepository: loginHistoryRepository,
	}
}

// GetUserLoginHistory получает историю входов пользователя
func (s *LoginHistoryService) GetUserLoginHistory(ctx context.Context, userID int64, limit, offset int) ([]dto.LoginHistoryDTO, error) {
	history, err := s.loginHistoryRepository.GetAllByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории входов: %w", err)
	}

	// Преобразуем в DTO
	result := make([]dto.LoginHistoryDTO, len(history))
	for i, entry := range history {
		result[i] = dto.LoginHistoryDTO{
			ID:        entry.ID,
			IPAddress: entry.IPAddress,
			UserAgent: entry.UserAgent,
			CreatedAt: entry.CreatedAt,
		}
	}

	return result, nil
}

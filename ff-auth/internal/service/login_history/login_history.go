package login_history

import (
	"context"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/postgres"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
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
func (s *LoginHistoryService) GetUserLoginHistory(ctx context.Context, userID int64, limit, offset int) ([]service.LoginHistoryParams, error) {
	history, err := s.loginHistoryRepository.GetAllByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории входов: %w", err)
	}

	// Преобразуем в параметры истории входов
	result := make([]service.LoginHistoryParams, len(history))
	for i, entry := range history {
		var userAgent *string
		if entry.UserAgent != "" {
			userAgent = &entry.UserAgent
		}
		result[i] = service.LoginHistoryParams{
			Id:        entry.ID,
			IpAddress: entry.IPAddress,
			UserAgent: userAgent,
			CreatedAt: entry.CreatedAt,
		}
	}

	return result, nil
}

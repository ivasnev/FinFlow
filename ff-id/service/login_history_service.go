package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-id/interfaces"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// LoginHistoryService реализует интерфейс LoginHistoryService
type LoginHistoryService struct {
	loginHistoryRepository interfaces.LoginHistoryRepository
}

// NewLoginHistoryService создает новый сервис истории входов
func NewLoginHistoryService(loginHistoryRepository interfaces.LoginHistoryRepository) *LoginHistoryService {
	return &LoginHistoryService{
		loginHistoryRepository: loginHistoryRepository,
	}
}

// GetByUserID получает историю входов пользователя
func (s *LoginHistoryService) GetByUserID(ctx context.Context, userID int64, page, pageSize int) ([]models.LoginHistory, error) {
	// Заглушка
	offset := (page - 1) * pageSize
	return s.loginHistoryRepository.GetAllByUserID(ctx, userID, pageSize, offset)
}

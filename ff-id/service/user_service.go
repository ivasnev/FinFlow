package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/interfaces"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// UserService реализует интерфейс UserService
type UserService struct {
	userRepository   interfaces.UserRepository
	avatarRepository interfaces.AvatarRepository
}

// NewUserService создает новый сервис пользователей
func NewUserService(
	userRepository interfaces.UserRepository,
	avatarRepository interfaces.AvatarRepository,
) *UserService {
	return &UserService{
		userRepository:   userRepository,
		avatarRepository: avatarRepository,
	}
}

// GetByID получает пользователя по ID
func (s *UserService) GetByID(ctx context.Context, id int64) (*models.User, error) {
	// Заглушка
	return &models.User{}, nil
}

// GetByNickname получает пользователя по никнейму
func (s *UserService) GetByNickname(ctx context.Context, nickname string) (*models.User, error) {
	// Заглушка
	return &models.User{}, nil
}

// Update обновляет данные пользователя
func (s *UserService) Update(ctx context.Context, id int64, userData *models.UserUpdate) (*models.User, error) {
	// Заглушка
	return &models.User{}, nil
}

// UpdateAvatar обновляет аватарку пользователя
func (s *UserService) UpdateAvatar(ctx context.Context, userID int64, avatarID uuid.UUID) error {
	// Заглушка
	return nil
}

// DeleteAvatar удаляет аватарку пользователя
func (s *UserService) DeleteAvatar(ctx context.Context, userID int64, avatarID uuid.UUID) error {
	// Заглушка
	return nil
}

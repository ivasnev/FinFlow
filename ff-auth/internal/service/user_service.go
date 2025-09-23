package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/postgres"

	"github.com/ivasnev/FinFlow/ff-auth/internal/api/dto"
	"golang.org/x/crypto/bcrypt"
)

// UserService реализует интерфейс для работы с пользователями
type UserService struct {
	userRepository postgres.UserRepositoryInterface
}

// NewUserService создает новый сервис пользователей
func NewUserService(
	userRepository postgres.UserRepositoryInterface,
) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

// GetUserByID получает пользователя по ID
func (s *UserService) GetUserByID(ctx context.Context, id int64) (*dto.UserDTO, error) {
	user, err := s.userRepository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	// Получаем роли пользователя
	roles, err := s.userRepository.GetRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения ролей пользователя: %w", err)
	}

	// Преобразуем роли в строки
	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = role.Name
	}

	// Формируем DTO для пользователя
	userDTO := &dto.UserDTO{
		ID:        user.ID,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Roles:     roleStrings,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	return userDTO, nil
}

// GetUserByNickname получает пользователя по никнейму
func (s *UserService) GetUserByNickname(ctx context.Context, nickname string) (*dto.UserDTO, error) {
	user, err := s.userRepository.GetByNickname(ctx, nickname)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	// Получаем роли пользователя
	roles, err := s.userRepository.GetRoles(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения ролей пользователя: %w", err)
	}

	// Преобразуем роли в строки
	roleStrings := make([]string, len(roles))
	for i, role := range roles {
		roleStrings[i] = role.Name
	}

	// Формируем DTO для пользователя
	userDTO := &dto.UserDTO{
		ID:        user.ID,
		Email:     user.Email,
		Nickname:  user.Nickname,
		Roles:     roleStrings,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return userDTO, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(ctx context.Context, userID int64, req dto.UpdateUserRequest) (*dto.UserDTO, error) {
	// Получаем пользователя по ID
	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	// Обновляем email, если указан
	if req.Email != nil {
		// Проверяем, не занят ли email другим пользователем
		if *req.Email != user.Email {
			existingUser, err := s.userRepository.GetByEmail(ctx, *req.Email)
			if err == nil && existingUser != nil && existingUser.ID != user.ID {
				return nil, errors.New("указанный email уже используется")
			}
			user.Email = *req.Email
		}
	}

	// Обновляем никнейм, если указан
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}

	// Обновляем пароль, если указан
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("ошибка хеширования пароля: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	}

	// Обновляем время изменения
	user.UpdatedAt = time.Now()

	// Сохраняем изменения
	if err := s.userRepository.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	// Получаем обновленного пользователя
	return s.GetUserByID(ctx, user.ID)
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(ctx context.Context, userID int64) error {
	return s.userRepository.Delete(ctx, userID)
}

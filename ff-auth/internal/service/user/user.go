package user

import (
	"context"
	"fmt"
	"time"

	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service/password"
	"golang.org/x/crypto/bcrypt"
)

// UserService реализует интерфейс для работы с пользователями
type UserService struct {
	userRepository    repository.User
	passwordMinLength int
	identity          service.Identity
}

// NewUserService создает новый сервис пользователей
func NewUserService(
	userRepository repository.User,
	passwordMinLength int,
	identity service.Identity,
) *UserService {
	return &UserService{
		userRepository:    userRepository,
		passwordMinLength: passwordMinLength,
		identity:          identity,
	}
}

// GetUserByID получает пользователя по ID
func (s *UserService) GetUserByID(ctx context.Context, id int64) (*service.UserData, error) {
	user, err := s.userRepository.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	identity, err := s.identity.GetIdentity(ctx, id)
	if err != nil {
		return nil, err
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

	updatedAt := user.UpdatedAt
	if profileUpdatedAt := time.Unix(identity.UpdatedAt, 0); profileUpdatedAt.After(updatedAt) {
		updatedAt = profileUpdatedAt
	}

	// Формируем данные для пользователя
	userData := &service.UserData{
		Id:        user.ID,
		Email:     identity.Email,
		Nickname:  identity.Nickname,
		Roles:     roleStrings,
		CreatedAt: user.CreatedAt,
		UpdatedAt: updatedAt,
	}
	return userData, nil
}

// GetUserByNickname получает пользователя по никнейму
func (s *UserService) GetUserByNickname(ctx context.Context, nickname string) (*service.UserData, error) {
	identity, err := s.identity.FindIdentity(ctx, nickname, ffid.LookupNickname)
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(ctx, identity.ID)
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(ctx context.Context, userID int64, req service.UserUpdateData) (*service.UserData, error) {
	if req.Password != nil {
		if err := password.Validate(*req.Password, s.passwordMinLength); err != nil {
			return nil, err
		}
	}

	// Получаем пользователя по ID
	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	// Хеш готовится до внешнего изменения; при отказе ff-id пароль остаётся прежним.
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("ошибка хеширования пароля: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	}
	var identity *ffid.UserDTO
	if req.Email != nil || req.Nickname != nil {
		identity, err = s.identity.UpdateIdentity(ctx, userID, req.Email, req.Nickname)
	} else {
		identity, err = s.identity.GetIdentity(ctx, userID)
	}
	if err != nil {
		return nil, err
	}

	// Пароль изменяет только время учётных данных. Время профиля принадлежит ff-id.
	if req.Password != nil {
		user.UpdatedAt = time.Now()
		if err := s.userRepository.Update(ctx, user); err != nil {
			return nil, fmt.Errorf("профиль доступен в ff-id, но пароль не сохранён; повторите смену пароля: %w", err)
		}
	}
	roles, err := s.userRepository.GetRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}
	updatedAt := user.UpdatedAt
	if profileUpdatedAt := time.Unix(identity.UpdatedAt, 0); profileUpdatedAt.After(updatedAt) {
		updatedAt = profileUpdatedAt
	}
	return &service.UserData{Id: userID, Email: identity.Email, Nickname: identity.Nickname, Roles: roleNames, CreatedAt: user.CreatedAt, UpdatedAt: updatedAt}, nil
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(ctx context.Context, userID int64) error {
	return s.userRepository.Delete(ctx, userID)
}

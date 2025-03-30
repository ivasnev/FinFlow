package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/dto"
	"github.com/ivasnev/FinFlow/ff-id/interfaces"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserService реализует интерфейс для работы с пользователями
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

	if user.Phone.Valid {
		phone := user.Phone.String
		userDTO.Phone = &phone
	}

	if user.Name.Valid {
		name := user.Name.String
		userDTO.Name = &name
	}

	if user.Birthdate.Valid {
		birthdate := user.Birthdate.Time
		userDTO.Birthdate = &birthdate
	}

	if user.AvatarID.Valid {
		avatarID := user.AvatarID.UUID
		userDTO.AvatarID = &avatarID
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

	if user.Phone.Valid {
		phone := user.Phone.String
		userDTO.Phone = &phone
	}

	if user.Name.Valid {
		name := user.Name.String
		userDTO.Name = &name
	}

	if user.Birthdate.Valid {
		birthdate := user.Birthdate.Time
		userDTO.Birthdate = &birthdate
	}

	if user.AvatarID.Valid {
		avatarID := user.AvatarID.UUID
		userDTO.AvatarID = &avatarID
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

	// Обновляем телефон, если указан
	if req.Phone != nil {
		user.Phone.String = *req.Phone
		user.Phone.Valid = true
	}

	// Обновляем имя, если указано
	if req.Name != nil {
		user.Name.String = *req.Name
		user.Name.Valid = true
	}

	// Обновляем дату рождения, если указана
	if req.Birthdate != nil {
		user.Birthdate.Time = *req.Birthdate
		user.Birthdate.Valid = true
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

// ChangeAvatar изменяет аватар пользователя
func (s *UserService) ChangeAvatar(ctx context.Context, userID int64, fileID uuid.UUID) error {
	// Получаем пользователя по ID
	user, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("ошибка получения пользователя: %w", err)
	}

	// Создаем новую аватарку
	avatarID := uuid.New()
	avatar := &models.UserAvatar{
		ID:         avatarID,
		UserID:     user.ID,
		FileID:     fileID,
		UploadedAt: time.Now(),
	}

	if err := s.avatarRepository.Create(ctx, avatar); err != nil {
		return fmt.Errorf("ошибка создания аватарки: %w", err)
	}

	// Обновляем пользователя
	user.AvatarID.UUID = avatarID
	user.AvatarID.Valid = true
	user.UpdatedAt = time.Now()

	if err := s.userRepository.Update(ctx, user); err != nil {
		return fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	return nil
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(ctx context.Context, userID int64) error {
	return s.userRepository.Delete(ctx, userID)
}

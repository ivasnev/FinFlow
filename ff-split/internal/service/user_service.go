package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"net/http"
	"strings"

	idclient "github.com/ivasnev/FinFlow/ff-id/pkg/client"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/slices"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"gorm.io/gorm"
)

// UserService реализует интерфейс для работы с пользователями
type UserService struct {
	db             *gorm.DB
	userRepository repository.UserRepositoryInterface
	idClient       *idclient.Client
}

// NewUserService создает новый сервис пользователей
func NewUserService(
	userRepository repository.UserRepositoryInterface,
	idClient *idclient.Client,
) *UserService {
	return &UserService{
		userRepository: userRepository,
		idClient:       idClient,
	}
}

// CreateUser создает нового пользователя
func (s *UserService) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	existingUser, err := s.userRepository.GetByExternalUserID(ctx, *user.UserID)
	if err == nil && existingUser != nil {
		return existingUser, nil // Пользователь уже существует, возвращаем его
	}

	// Создаем пользователя
	createdUser, err := s.userRepository.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	return createdUser, nil
}

// CreateDummyUser создает нового dummy-пользователя для мероприятия
func (s *UserService) CreateDummyUser(ctx context.Context, name string, eventID int64) (*models.User, error) {
	// Создаем dummy пользователя без UserID (он будет заполнен автоматически)
	dummyUser := &models.User{
		NameCashed: name,
		IsDummy:    true,
	}

	// Создаем пользователя
	createdUser, err := s.userRepository.Create(ctx, dummyUser)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании dummy пользователя: %w", err)
	}

	// Добавляем пользователя в мероприятие
	if err := s.AddUserToEvent(ctx, createdUser.ID, eventID); err != nil {
		return nil, fmt.Errorf("ошибка при добавлении dummy пользователя в мероприятие: %w", err)
	}

	return createdUser, nil
}

// BatchCreateDummyUsers создает dummy-пользователей для мероприятия
func (s *UserService) BatchCreateDummyUsers(ctx context.Context, names []string, eventID int64) ([]*models.User, error) {
	users := make([]*models.User, len(names))

	for i, name := range names {
		users[i] = &models.User{
			NameCashed: name,
			IsDummy:    true,
		}
	}

	if err := s.userRepository.BatchCreate(ctx, users); err != nil {
		return nil, fmt.Errorf("ошибка при создании dummy пользователей: %w", err)
	}

	return users, nil
}

// GetUserByID получает пользователя по внутреннему ID
func (s *UserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	user, err := s.userRepository.GetByInternalUserID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}
	return user, nil
}

func (s *UserService) GetUsersByExternalUserIDs(ctx context.Context, userIDs []int64) ([]models.User, error) {
	users, err := s.userRepository.GetByExternalUserIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователей: %w", err)
	}
	return users, nil
}

func (s *UserService) GetUsersByInternalUserIDs(ctx context.Context, userIDs []int64) ([]models.User, error) {
	users, err := s.userRepository.GetByInternalUserIDs(ctx, userIDs)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователей: %w", err)
	}
	return users, nil
}

// GetUserByUserID получает пользователя по UserID (ID из сервиса идентификации)
func (s *UserService) GetUserByUserID(ctx context.Context, userID int64) (*models.User, error) {
	// Пытаемся найти пользователя в базе данных
	user, err := s.userRepository.GetByExternalUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}

	return user, nil
}

// GetUsersByEventID получает всех пользователей мероприятия
func (s *UserService) GetUsersByEventID(ctx context.Context, eventID int64) ([]models.User, error) {
	// Сначала получаем всех пользователей из локальной базы
	users, err := s.userRepository.GetByEventID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователей мероприятия: %w", err)
	}

	return users, nil
}

// GetDummiesByEventID получает всех dummy-пользователей мероприятия
func (s *UserService) GetDummiesByEventID(ctx context.Context, eventID int64) ([]models.User, error) {
	users, err := s.userRepository.GetDummiesByEventID(ctx, eventID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении dummy-пользователей мероприятия: %w", err)
	}
	return users, nil
}

// UpdateUser обновляет данные пользователя
func (s *UserService) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	// Получаем существующего пользователя
	existingUser, err := s.userRepository.GetByExternalUserID(ctx, *user.UserID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}

	// Если это не dummy пользователь, не позволяем обновлять кэшированные данные напрямую
	if !existingUser.IsDummy {
		return nil, errors.New("невозможно обновить кэшированные данные для не-dummy пользователя")
	}

	// Обновляем только разрешенные поля для dummy пользователя
	existingUser.NameCashed = user.NameCashed

	// Обновляем пользователя
	updatedUser, err := s.userRepository.Update(ctx, existingUser)
	if err != nil {
		return nil, fmt.Errorf("ошибка при обновлении пользователя: %w", err)
	}

	return updatedUser, nil
}

// DeleteUser удаляет пользователя
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {
	if err := s.userRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("ошибка при удалении пользователя: %w", err)
	}
	return nil
}

// AddUsersToEvent добавляет пользователей в мероприятие
func (s *UserService) AddUsersToEvent(ctx context.Context, ids []int64, eventID int64) error {
	internalUser, err := s.userRepository.GetByInternalUserIDs(ctx, ids)
	if err != nil {
		return fmt.Errorf("ошибка при получении пользователей: %w", err)
	}

	internalUserIds := make([]int64, len(internalUser))
	for i, user := range internalUser {
		internalUserIds[i] = user.ID
	}

	if err := s.userRepository.AddUsersToEvent(ctx, internalUserIds, eventID); err != nil {
		return fmt.Errorf("ошибка при добавлении пользователей в мероприятие: %w", err)
	}

	return nil
}

// AddUserToEvent добавляет пользователя в мероприятие
func (s *UserService) AddUserToEvent(ctx context.Context, idUser, eventID int64) error {

	if err := s.userRepository.AddUserToEvent(ctx, idUser, eventID); err != nil {
		return fmt.Errorf("ошибка при добавлении пользователя в мероприятие: %w", err)
	}

	return nil
}

// RemoveUserFromEvent удаляет пользователя из мероприятия
func (s *UserService) RemoveUserFromEvent(ctx context.Context, userID, eventID int64) error {
	if err := s.userRepository.RemoveUserFromEvent(ctx, userID, eventID); err != nil {
		return fmt.Errorf("ошибка при удалении пользователя из мероприятия: %w", err)
	}
	return nil
}

// SyncUserWithIDService синхронизирует данные пользователя с ID-сервисом
func (s *UserService) SyncUserWithIDService(ctx context.Context, userID int64) (*models.User, error) {
	if userID <= 0 {
		return nil, errors.New("невозможно синхронизировать dummy пользователя с ID-сервисом")
	}

	// Получаем информацию о пользователе из ID-сервиса
	resp, err := s.idClient.ApiRequest(http.MethodGet, fmt.Sprintf("/internal/users?user_id=%d", userID), nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе к ID-сервису: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ID-сервис вернул код ошибки: %d", resp.StatusCode)
	}

	// Парсим ответ
	var userInfoResponse struct {
		Users []struct {
			ID    int64  `json:"id"`
			Name  string `json:"name"`
			Photo string `json:"photo,omitempty"`
			Email string `json:"email,omitempty"`
		} `json:"users"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfoResponse); err != nil {
		return nil, fmt.Errorf("ошибка при парсинге ответа от ID-сервиса: %w", err)
	}

	// Проверяем, что пользователь найден
	if len(userInfoResponse.Users) == 0 {
		return nil, errors.New("пользователь не найден в ID-сервисе")
	}

	// Получаем первого пользователя из ответа
	userInfo := userInfoResponse.Users[0]

	user := &models.User{
		UserID:          &userInfo.ID,
		NameCashed:      userInfo.Name,
		PhotoUUIDCashed: userInfo.Photo,
		IsDummy:         false,
	}

	if err := s.userRepository.CreateOrUpdate(ctx, user); err != nil {
		return nil, fmt.Errorf("ошибка при создании или обновлении пользователя: %w", err)
	}

	return user, nil
}

// BatchSyncUsersWithIDService синхронизирует данные группы пользователей с ID-сервисом
func (s *UserService) BatchSyncUsersWithIDService(ctx context.Context, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}
	queryParams := make([]string, len(userIDs))
	for i, userID := range userIDs {
		queryParams[i] = fmt.Sprintf("user_id=%d", userID)
	}

	// Формируем запрос к ID-сервису
	resp, err := s.idClient.ApiRequest(http.MethodGet, "/internal/users?"+strings.Join(queryParams, "&"), nil)
	if err != nil {
		return fmt.Errorf("ошибка при запросе к ID-сервису: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ID-сервис вернул код ошибки: %d", resp.StatusCode)
	}

	// Парсим ответ
	var userInfoResponse dto.ResponseFromIDService

	if err := json.NewDecoder(resp.Body).Decode(&userInfoResponse); err != nil {
		return fmt.Errorf("ошибка при парсинге ответа от ID-сервиса: %w", err)
	}

	users := make([]*models.User, len(userInfoResponse))
	for i, userInfo := range userInfoResponse {
		users[i] = &models.User{
			UserID:         &userInfo.ID,
			NicknameCashed: userInfo.Nickname,
			IsDummy:        false,
		}
		if userInfo.AvatarID != nil {
			users[i].PhotoUUIDCashed = userInfo.AvatarID.String()
		}
		if userInfo.Name != nil {
			users[i].NameCashed = *userInfo.Name
		}
	}

	if err := s.userRepository.BatchCreateOrUpdate(ctx, users); err != nil {
		return fmt.Errorf("ошибка при создании или обновлении пользователей: %w", err)
	}

	return nil
}

func (s *UserService) GetNotExistsUsers(ctx context.Context, ids []int64) ([]int64, error) {
	users, err := s.userRepository.GetByExternalUserIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователей: %w", err)
	}
	var existsUsersIds []int64
	for _, user := range users {
		existsUsersIds = append(existsUsersIds, *user.UserID)
	}
	notExistsUsers := slices.SmartDiff(ids, existsUsersIds)

	return notExistsUsers, nil
}

package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/ivasnev/FinFlow/ff-split/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/slices"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	"gorm.io/gorm"
)

// UserService реализует интерфейс для работы с пользователями
type UserService struct {
	db             *gorm.DB
	userRepository repository.User
	idAdapter      *ffid.Adapter
}

// NewUserService создает новый сервис пользователей
func NewUserService(
	userRepository repository.User,
	idAdapter *ffid.Adapter,
) *UserService {
	return &UserService{
		userRepository: userRepository,
		idAdapter:      idAdapter,
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

// GetUsersByExternalIDs возвращает внутренние ID пользователей по внешним ID
// Если пользователь не найден в системе, синхронизирует его с ff-id сервисом
func (s *UserService) GetUsersByExternalIDs(ctx context.Context, externalIDs []int64) ([]service.ExternalToInternalMapping, error) {
	mappings := make([]service.ExternalToInternalMapping, 0, len(externalIDs))

	for _, externalID := range externalIDs {
		// Сначала пытаемся найти пользователя в локальной базе
		user, err := s.userRepository.GetByExternalUserID(ctx, externalID)
		if err == nil && user != nil {
			// Пользователь найден в локальной базе
			mappings = append(mappings, service.ExternalToInternalMapping{
				ExternalID:  externalID,
				InternalID:  user.ID,
				UserProfile: user,
			})
			continue
		}

		// Пользователь не найден, синхронизируем с ff-id сервисом
		syncedUser, err := s.SyncUserWithIDService(ctx, externalID)
		if err != nil {
			return nil, fmt.Errorf("ошибка при синхронизации пользователя %d: %w", externalID, err)
		}

		mappings = append(mappings, service.ExternalToInternalMapping{
			ExternalID:  externalID,
			InternalID:  syncedUser.ID,
			UserProfile: syncedUser,
		})
	}

	return mappings, nil
}

// GetUserByInternalUserID получает пользователя по внутреннему ID
func (s *UserService) GetUserByInternalUserID(ctx context.Context, id int64) (*models.User, error) {
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

// GetUserByExternalUserID получает пользователя по UserID (ID из сервиса идентификации)
func (s *UserService) GetUserByExternalUserID(ctx context.Context, userID int64) (*models.User, error) {
	// Пытаемся найти пользователя в базе данных
	user, err := s.userRepository.GetByExternalUserID(ctx, userID)
	if err == nil {
		// Пользователь найден
		return user, nil
	}

	// Проверяем, что это именно ошибка "не найден"
	if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}

	// Пользователь не найден, синхронизируем с ID-сервисом
	syncedUser, syncErr := s.SyncUserWithIDService(ctx, userID)
	if syncErr != nil {
		return nil, fmt.Errorf("ошибка при синхронизации пользователя: %w", syncErr)
	}

	return syncedUser, nil
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

func (s *UserService) GetInternalUserIdsByExternalUserIds(ctx context.Context, externalUserIds []int64) ([]int64, error) {
	internalUser, err := s.userRepository.GetByExternalUserIDs(ctx, externalUserIds)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователей: %w", err)
	}
	if len(internalUser) != len(externalUserIds) {
		return nil, errors.New("некоторые пользователи не найдены")
	}

	internalUserIds := make([]int64, len(internalUser))
	for i, user := range internalUser {
		internalUserIds[i] = user.ID
	}
	return internalUserIds, nil
}

// AddUsersToEvent добавляет пользователей в мероприятие
func (s *UserService) AddUsersToEvent(ctx context.Context, internalUserIds []int64, eventID int64) error {
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

	// Получаем информацию о пользователе из ID-сервиса через адаптер
	userInfo, err := s.idAdapter.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка при запросе к ID-сервису: %w", err)
	}

	user := &models.User{
		UserID:          &userInfo.ID,
		NicknameCashed:  userInfo.Nickname,
		NameCashed:      "",
		PhotoUUIDCashed: "",
		IsDummy:         false,
	}

	if userInfo.Name != nil {
		user.NameCashed = *userInfo.Name
	}

	if userInfo.AvatarID != nil {
		user.PhotoUUIDCashed = *userInfo.AvatarID
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

	// Получаем информацию о пользователях из ID-сервиса через адаптер
	usersInfo, err := s.idAdapter.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return fmt.Errorf("ошибка при запросе к ID-сервису: %w", err)
	}

	users := make([]*models.User, len(usersInfo))
	for i, userInfo := range usersInfo {
		users[i] = &models.User{
			UserID:         &userInfo.ID,
			NicknameCashed: userInfo.Nickname,
			IsDummy:        false,
		}
		if userInfo.AvatarID != nil {
			users[i].PhotoUUIDCashed = *userInfo.AvatarID
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

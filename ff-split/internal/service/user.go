package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// User определяет методы для работы с пользователями
type User interface {
	// CreateUser создает нового пользователя
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)

	// CreateDummyUser создает нового dummy-пользователя для мероприятия
	CreateDummyUser(ctx context.Context, name string, eventID int64) (*models.User, error)

	// BatchCreateDummyUsers создает dummy-пользователей для мероприятия
	BatchCreateDummyUsers(ctx context.Context, names []string, eventID int64) ([]*models.User, error)

	// GetUserByInternalUserID получает пользователя по внутреннему ID
	GetUserByInternalUserID(ctx context.Context, id int64) (*models.User, error)

	// GetUserByExternalUserID получает пользователя по UserID (ID из сервиса идентификации)
	GetUserByExternalUserID(ctx context.Context, userID int64) (*models.User, error)

	// GetUsersByExternalUserIDs получает пользователей по UserID (ID из сервиса идентификации)
	GetUsersByExternalUserIDs(ctx context.Context, userIDs []int64) ([]models.User, error)

	// GetUsersByInternalUserIDs получает пользователей по UserID (ID из сервиса идентификации)
	GetUsersByInternalUserIDs(ctx context.Context, userIDs []int64) ([]models.User, error)

	// GetInternalUserIdsByExternalUserIds получает внутренние ID пользователей по внешним ID
	GetInternalUserIdsByExternalUserIds(ctx context.Context, externalUserIds []int64) ([]int64, error)

	// GetUsersByEventID получает всех пользователей мероприятия
	GetUsersByEventID(ctx context.Context, eventID int64) ([]models.User, error)

	// GetDummiesByEventID получает всех dummy-пользователей мероприятия
	GetDummiesByEventID(ctx context.Context, eventID int64) ([]models.User, error)

	// UpdateUser обновляет данные пользователя
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)

	// DeleteUser удаляет пользователя
	DeleteUser(ctx context.Context, id int64) error

	// AddUsersToEvent добавляет пользователей в мероприятие
	AddUsersToEvent(ctx context.Context, ids []int64, eventID int64) error

	// RemoveUserFromEvent удаляет пользователя из мероприятия
	RemoveUserFromEvent(ctx context.Context, userID, eventID int64) error

	// SyncUserWithIDService синхронизирует данные пользователя с ID-сервисом
	SyncUserWithIDService(ctx context.Context, userID int64) (*models.User, error)

	// BatchSyncUsersWithIDService синхронизирует данные пользователей с ID-сервисом
	BatchSyncUsersWithIDService(ctx context.Context, userIDs []int64) error

	// GetNotExistsUsers получает пользователей, которые не существуют в базе данных
	GetNotExistsUsers(ctx context.Context, ids []int64) ([]int64, error)
}


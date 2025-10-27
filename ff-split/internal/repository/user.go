package repository

import (
	"context"
	"errors"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

var ErrUserNotFound = errors.New("пользователь не найден")

// User определяет методы для работы с пользователями в репозитории
type User interface {
	// Create создает нового пользователя
	Create(ctx context.Context, user *models.User) (*models.User, error)

	// BatchCreate создает пользователей
	BatchCreate(ctx context.Context, users []*models.User) error

	// BatchCreateOrUpdate создает или обновляет пользователей
	BatchCreateOrUpdate(ctx context.Context, users []*models.User) error

	// CreateOrUpdate создает или обновляет пользователя
	CreateOrUpdate(ctx context.Context, user *models.User) error

	// GetByExternalUserIDs находит пользователей по UserID (ID из сервиса идентификации)
	GetByExternalUserIDs(ctx context.Context, ids []int64) ([]models.User, error)

	// GetByInternalUserIDs находит пользователей по внутренним ID
	GetByInternalUserIDs(ctx context.Context, ids []int64) ([]models.User, error)

	// GetByInternalUserID находит пользователя по внутреннему ID
	GetByInternalUserID(ctx context.Context, id int64) (*models.User, error)

	// GetByExternalUserID находит пользователя по UserID (ID из сервиса идентификации)
	GetByExternalUserID(ctx context.Context, userID int64) (*models.User, error)

	// GetByEventID находит всех пользователей, связанных с мероприятием
	GetByEventID(ctx context.Context, eventID int64) ([]models.User, error)

	// GetDummiesByEventID находит всех dummy-пользователей, связанных с мероприятием
	GetDummiesByEventID(ctx context.Context, eventID int64) ([]models.User, error)

	// Update обновляет данные пользователя
	Update(ctx context.Context, user *models.User) (*models.User, error)

	// Delete удаляет пользователя
	Delete(ctx context.Context, id int64) error

	// AddUserToEvent добавляет пользователя в мероприятие
	AddUserToEvent(ctx context.Context, userID, eventID int64) error

	// AddUsersToEvent добавляет пользователей в мероприятие
	AddUsersToEvent(ctx context.Context, ids []int64, eventID int64) error

	// RemoveUserFromEvent удаляет пользователя из мероприятия
	RemoveUserFromEvent(ctx context.Context, userID, eventID int64) error
}

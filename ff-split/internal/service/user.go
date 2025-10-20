package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// UserFromId представляет данные пользователя из ID сервиса
type UserFromId struct {
	ID        int64      `json:"id"`
	Email     string     `json:"email"`
	Phone     *string    `json:"phone,omitempty"`
	Nickname  string     `json:"nickname"`
	Name      *string    `json:"name,omitempty"`
	Birthdate *int64     `json:"birthdate,omitempty"`
	AvatarID  *uuid.UUID `json:"avatar_id,omitempty"`
	CreatedAt int64      `json:"created_at"`
	UpdatedAt int64      `json:"updated_at"`
}

// ResponseFromIDService представляет ответ от ID сервиса
type ResponseFromIDService []UserFromId

// CreateUserRequest представляет запрос на создание пользователя
type CreateUserRequest struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Name     string `json:"name"`
	Photo    string `json:"photo"`
}

// UpdateUserProfileDTO представляет запрос на обновление профиля пользователя
type UpdateUserProfileDTO struct {
	UserID   int64   `json:"user_id"`
	Nickname *string `json:"nickname"`
	Name     *string `json:"name"`
	Photo    *string `json:"photo"`
}

// UserProfileDTO представляет профиль пользователя
type UserProfileDTO struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Name     string `json:"name"`
	Photo    string `json:"photo"`
}

// UserResponse представляет ответ с данными пользователя
type UserResponse struct {
	ID      int64           `json:"id"`
	Name    string          `json:"name"`
	IsDummy bool            `json:"is_dummy"`
	Profile *UserProfileDTO `json:"profile,omitempty"`
}

// CreateUserResponse представляет ответ на создание пользователя
type CreateUserResponse struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Name     string `json:"name"`
	Photo    string `json:"photo"`
}

// GetUserResponse представляет ответ на получение пользователя
type GetUserResponse struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Name     string `json:"name"`
	Photo    string `json:"photo"`
}

// GetUsersResponse представляет ответ на получение списка пользователей
type GetUsersResponse []GetUserResponse

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

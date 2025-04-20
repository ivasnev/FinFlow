package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// UserRepositoryInterface определяет методы для работы с пользователями
type UserRepositoryInterface interface {
	// Create создает нового пользователя
	Create(ctx context.Context, user *models.User) error

	// GetByID получает пользователя по ID
	GetByID(ctx context.Context, id int64) (*models.User, error)

	// GetByEmail получает пользователя по email
	GetByEmail(ctx context.Context, email string) (*models.User, error)

	// GetByNickname получает пользователя по никнейму
	GetByNickname(ctx context.Context, nickname string) (*models.User, error)

	// Update обновляет данные пользователя
	Update(ctx context.Context, user *models.User) error

	// Delete удаляет пользователя
	Delete(ctx context.Context, id int64) error
}

// AvatarRepositoryInterface определяет методы для работы с аватарами
type AvatarRepositoryInterface interface {
	// Create создает новый аватар
	Create(ctx context.Context, avatar *models.UserAvatar) error

	// GetByID получает аватар по ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.UserAvatar, error)

	// GetAllByUserID получает все аватарки пользователя
	GetAllByUserID(ctx context.Context, userID int64) ([]models.UserAvatar, error)

	// Delete удаляет аватар
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteAllByUserID удаляет все аватарки пользователя
	DeleteAllByUserID(ctx context.Context, userID int64) error
}

// FriendRepositoryInterface определяет методы для работы с друзьями пользователей
type FriendRepositoryInterface interface {
	// AddFriend добавляет пользователя в друзья (создает заявку)
	AddFriend(ctx context.Context, userID, friendID int64) error

	// UpdateFriendStatus обновляет статус дружбы
	UpdateFriendStatus(ctx context.Context, userID, friendID int64, status string) error

	// CreateMutualFriendship создает взаимную дружбу (при принятии заявки)
	CreateMutualFriendship(ctx context.Context, userID, friendID int64) error

	// RemoveFriend удаляет пользователя из друзей
	RemoveFriend(ctx context.Context, userID, friendID int64) error

	// GetFriendRelation получает информацию о связи дружбы между пользователями
	GetFriendRelation(ctx context.Context, userID, friendID int64) (*models.UserFriend, error)

	// GetFriendRelationWithPreload получает информацию о связи дружбы с предзагрузкой связей
	GetFriendRelationWithPreload(ctx context.Context, userID, friendID int64, preloadUser, preloadFriend bool) (*models.UserFriend, error)

	// GetFriends получает список друзей пользователя с пагинацией и фильтрацией
	GetFriends(ctx context.Context, userID int64, page, pageSize int, friendName, status string) ([]models.UserFriend, int64, error)

	// GetFriendRequests получает список заявок в друзья пользователя
	GetFriendRequests(ctx context.Context, userID int64, page, pageSize int, incoming bool) ([]models.UserFriend, int64, error)
}

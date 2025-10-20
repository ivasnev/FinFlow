package repository

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// Friend определяет методы для работы с друзьями пользователей
type Friend interface {
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

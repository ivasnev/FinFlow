package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/internal/api/dto"
)

// UserServiceInterface определяет методы для работы с пользователями
type UserServiceInterface interface {
	// GetUserByID получает пользователя по ID
	GetUserByID(ctx context.Context, id int64) (*dto.UserDTO, error)

	// GetUserByNickname получает пользователя по никнейму
	GetUserByNickname(ctx context.Context, nickname string) (*dto.UserDTO, error)

	// UpdateUser обновляет данные пользователя
	UpdateUser(ctx context.Context, userID int64, req dto.UpdateUserRequest) (*dto.UserDTO, error)

	// ChangeAvatar изменяет аватар пользователя
	ChangeAvatar(ctx context.Context, userID int64, fileID uuid.UUID) error

	// DeleteUser удаляет пользователя
	DeleteUser(ctx context.Context, userID int64) error

	// RegisterUser регистрирует нового пользователя
	RegisterUser(ctx context.Context, userID int64, user *dto.RegisterUserRequest) (*dto.UserDTO, error)
}

// FriendServiceInterface определяет методы для работы с друзьями пользователей
type FriendServiceInterface interface {
	// AddFriend создает заявку на добавление в друзья
	AddFriend(ctx context.Context, userID int64, req dto.AddFriendRequest) error

	// AcceptFriendRequest принимает заявку в друзья
	AcceptFriendRequest(ctx context.Context, userID int64, req dto.FriendActionRequest) error

	// RejectFriendRequest отклоняет заявку в друзья
	RejectFriendRequest(ctx context.Context, userID int64, req dto.FriendActionRequest) error

	// BlockUser блокирует пользователя
	BlockUser(ctx context.Context, userID int64, req dto.FriendActionRequest) error

	// RemoveFriend удаляет пользователя из друзей
	RemoveFriend(ctx context.Context, userID, friendID int64) error

	// GetFriendStatus получает статус дружбы между пользователями
	GetFriendStatus(ctx context.Context, userID, friendID int64) (string, error)

	// GetFriends получает список друзей пользователя
	GetFriends(ctx context.Context, nickname string, params dto.FriendsQueryParams) (*dto.FriendsListResponse, error)

	// GetFriendRequests получает список заявок в друзья
	GetFriendRequests(ctx context.Context, userID int64, page, pageSize int, incoming bool) (*dto.FriendsListResponse, error)
}

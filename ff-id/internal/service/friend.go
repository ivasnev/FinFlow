package service

import (
	"context"

	"github.com/google/uuid"
)

// Константы статусов дружбы
const (
	FriendStatusPending  = "pending"
	FriendStatusAccepted = "accepted"
	FriendStatusRejected = "rejected"
	FriendStatusBlocked  = "blocked"
)

// AddFriendRequest запрос на добавление друга
type AddFriendRequest struct {
	FriendNickname string
}

// FriendActionRequest запрос на действие с заявкой в друзья
type FriendActionRequest struct {
	UserID int64
	Action string
}

// FriendDTO представление друга пользователя
type FriendDTO struct {
	UserID  int64
	PhotoID uuid.UUID
	Name    string
	Status  string
}

// FriendsListResponse ответ на запрос списка друзей с пагинацией
type FriendsListResponse struct {
	Page     int
	PageSize int
	Total    int64
	Objects  []FriendDTO
}

// FriendsQueryParams параметры запроса для списка друзей
type FriendsQueryParams struct {
	Page       int
	PageSize   int
	FriendName string
	Status     string
}

// FriendServiceInterface определяет методы для работы с друзьями пользователей
type FriendServiceInterface interface {
	// AddFriend создает заявку на добавление в друзья
	AddFriend(ctx context.Context, userID int64, req AddFriendRequest) error

	// AcceptFriendRequest принимает заявку в друзья
	AcceptFriendRequest(ctx context.Context, userID int64, req FriendActionRequest) error

	// RejectFriendRequest отклоняет заявку в друзья
	RejectFriendRequest(ctx context.Context, userID int64, req FriendActionRequest) error

	// BlockUser блокирует пользователя
	BlockUser(ctx context.Context, userID int64, req FriendActionRequest) error

	// RemoveFriend удаляет пользователя из друзей
	RemoveFriend(ctx context.Context, userID, friendID int64) error

	// GetFriendStatus получает статус дружбы между пользователями
	GetFriendStatus(ctx context.Context, userID, friendID int64) (string, error)

	// GetFriends получает список друзей пользователя
	GetFriends(ctx context.Context, nickname string, params FriendsQueryParams) (*FriendsListResponse, error)

	// GetFriendRequests получает список заявок в друзья
	GetFriendRequests(ctx context.Context, userID int64, page, pageSize int, incoming bool) (*FriendsListResponse, error)
}

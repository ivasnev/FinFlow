package dto

import (
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
	FriendNickname string `json:"friend_nickname" binding:"required"`
}

// FriendActionRequest запрос на действие с заявкой в друзья
type FriendActionRequest struct {
	UserID int64  `json:"user_id" binding:"required"`
	Action string `json:"action" binding:"required,oneof=accept reject block"`
}

// FriendDTO представление друга пользователя
type FriendDTO struct {
	UserID  int64     `json:"user_id"`
	PhotoID uuid.UUID `json:"photo_id,omitempty"`
	Name    string    `json:"name"`
	Status  string    `json:"status,omitempty"`
}

// FriendsListResponse ответ на запрос списка друзей с пагинацией
type FriendsListResponse struct {
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int64       `json:"total"`
	Objects  []FriendDTO `json:"objects"`
}

// FriendsQueryParams параметры запроса для списка друзей
type FriendsQueryParams struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	FriendName string `form:"friend_name"`
	Status     string `form:"status"`
}

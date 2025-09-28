package dto

import "github.com/google/uuid"

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

type ResponseFromIDService []UserFromId

type CreateUserRequest struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Name     string `json:"name"`
	Photo    string `json:"photo"`
}

type UpdateUserProfileDTO struct {
	UserID   int64   `json:"user_id"`
	Nickname *string `json:"nickname"`
	Name     *string `json:"name"`
	Photo    *string `json:"photo"`
}

type UserProfileDTO struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Name     string `json:"name"`
	Photo    string `json:"photo"`
}

type UserResponse struct {
	ID      int64           `json:"id"`
	Name    string          `json:"name"`
	IsDummy bool            `json:"is_dummy"`
	Profile *UserProfileDTO `json:"profile,omitempty"`
}

type CreateUserResponse struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Name     string `json:"name"`
	Photo    string `json:"photo"`
}

type GetUserResponse struct {
	UserID   int64  `json:"user_id"`
	Nickname string `json:"nickname"`
	Name     string `json:"name"`
	Photo    string `json:"photo"`
}

type GetUsersResponse []GetUserResponse

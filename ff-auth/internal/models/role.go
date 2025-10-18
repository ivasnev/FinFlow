package models

// Role определяет роль пользователя в системе
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
)

// RoleEntity представляет роль в системе
type RoleEntity struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// UserRole представляет связь между пользователем и ролью
type UserRole struct {
	UserID int64 `json:"user_id"`
	RoleID int   `json:"role_id"`
}

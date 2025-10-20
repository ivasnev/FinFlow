package ffid

// RegisterUserRequest - запрос на регистрацию пользователя
type RegisterUserRequest struct {
	UserID   int64
	Email    string
	Nickname string
}

// UserDTO - данные пользователя
type UserDTO struct {
	ID        int64
	Email     string
	Nickname  string
	Name      *string
	Phone     *string
	AvatarID  *string
	Birthdate *int64
	CreatedAt int64
	UpdatedAt int64
}

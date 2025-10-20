package ffid

// UserDTO - данные пользователя из ff-id сервиса
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

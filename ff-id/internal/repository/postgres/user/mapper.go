package user

import (
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// ExtractUser преобразует модель пользователя базы данных в доменную модель
func ExtractUser(dbUser *User) *models.User {
	if dbUser == nil {
		return nil
	}

	return &models.User{
		ID:        dbUser.ID,
		Email:     dbUser.Email,
		Phone:     dbUser.Phone,
		Nickname:  dbUser.Nickname,
		Name:      dbUser.Name,
		Birthdate: dbUser.Birthdate,
		AvatarID:  dbUser.AvatarID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
	}
}

// LoadUser преобразует доменную модель пользователя в модель базы данных
func LoadUser(user *models.User) *User {
	if user == nil {
		return nil
	}

	return &User{
		ID:        user.ID,
		Email:     user.Email,
		Phone:     user.Phone,
		Nickname:  user.Nickname,
		Name:      user.Name,
		Birthdate: user.Birthdate,
		AvatarID:  user.AvatarID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

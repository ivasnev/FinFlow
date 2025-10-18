package user

import (
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// ExtractUser преобразует модель пользователя базы данных в обычную модель
func ExtractUser(dbUser *User) *models.User {
	if dbUser == nil {
		return nil
	}

	return &models.User{
		ID:           dbUser.ID,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		Nickname:     dbUser.Nickname,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}
}

// loadUser преобразует обычную модель пользователя в модель базы данных
func loadUser(user *models.User) *User {
	if user == nil {
		return nil
	}

	return &User{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Nickname:     user.Nickname,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

// ExtractRole преобразует модель роли базы данных в обычную модель
func ExtractRole(dbRole *RoleEntity) *models.RoleEntity {
	if dbRole == nil {
		return nil
	}

	return &models.RoleEntity{
		ID:   dbRole.ID,
		Name: dbRole.Name,
	}
}

// loadRole преобразует обычную модель роли в модель базы данных
func loadRole(role *models.RoleEntity) *RoleEntity {
	if role == nil {
		return nil
	}

	return &RoleEntity{
		ID:   role.ID,
		Name: role.Name,
	}
}

// ExtractUserRole преобразует модель связи пользователь-роль базы данных в обычную модель
func ExtractUserRole(dbUserRole *UserRole) *models.UserRole {
	if dbUserRole == nil {
		return nil
	}

	return &models.UserRole{
		UserID: dbUserRole.UserID,
		RoleID: dbUserRole.RoleID,
	}
}

// loadUserRole преобразует обычную модель связи пользователь-роль в модель базы данных
func loadUserRole(userRole *models.UserRole) *UserRole {
	if userRole == nil {
		return nil
	}

	return &UserRole{
		UserID: userRole.UserID,
		RoleID: userRole.RoleID,
	}
}

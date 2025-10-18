package role

import (
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

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

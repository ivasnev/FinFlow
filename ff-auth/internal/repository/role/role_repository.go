package role

import (
	"context"
	"errors"

	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository"
	"gorm.io/gorm"
)

// RoleRepository реализует интерфейс для работы с ролями в PostgreSQL через GORM
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository создает новый репозиторий ролей
func NewRoleRepository(db *gorm.DB) repository.Role {
	return &RoleRepository{
		db: db,
	}
}

// GetByName находит роль по имени
func (r *RoleRepository) GetByName(ctx context.Context, name string) (*models.RoleEntity, error) {
	var role RoleEntity
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("роль не найдена")
		}
		return nil, err
	}
	return ExtractRole(&role), nil
}

// GetAll получает все роли
func (r *RoleRepository) GetAll(ctx context.Context) ([]models.RoleEntity, error) {
	var roles []RoleEntity
	err := r.db.WithContext(ctx).Find(&roles).Error
	if err != nil {
		return nil, err
	}
	var roleModels []models.RoleEntity
	for _, role := range roles {
		roleModels = append(roleModels, *ExtractRole(&role))
	}
	return roleModels, nil
}

package postgres

import (
	"context"
	"errors"

	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"gorm.io/gorm"
)

// RoleRepository реализует интерфейс для работы с ролями в PostgreSQL через GORM
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository создает новый репозиторий ролей
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{
		db: db,
	}
}

// GetByName находит роль по имени
func (r *RoleRepository) GetByName(ctx context.Context, name string) (*models.RoleEntity, error) {
	var role models.RoleEntity
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("роль не найдена")
		}
		return nil, err
	}
	return &role, nil
}

// GetAll получает все роли
func (r *RoleRepository) GetAll(ctx context.Context) ([]models.RoleEntity, error) {
	var roles []models.RoleEntity
	err := r.db.WithContext(ctx).Find(&roles).Error
	return roles, err
}

package user

import (
	"context"
	"errors"

	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository"
	"gorm.io/gorm"
)

// UserRepository реализует интерфейс для работы с пользователями в PostgreSQL через GORM
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(db *gorm.DB) repository.User {
	return &UserRepository{
		db: db,
	}
}

// Create создает нового пользователя
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	dbUser := loadUser(user)
	if err := r.db.WithContext(ctx).Create(dbUser).Error; err != nil {
		return err
	}
	// Обновляем ID пользователя после создания
	user.ID = dbUser.ID
	return nil
}

// GetByID находит пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	var user User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}
	return ExtractUser(&user), nil
}

// GetByEmail находит пользователя по email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}
	return ExtractUser(&user), nil
}

// GetByNickname находит пользователя по никнейму
func (r *UserRepository) GetByNickname(ctx context.Context, nickname string) (*models.User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("nickname = ?", nickname).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}
	return ExtractUser(&user), nil
}

// Update обновляет данные пользователя
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	dbUser := loadUser(user)
	return r.db.WithContext(ctx).Save(dbUser).Error
}

// Delete удаляет пользователя
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

// AddRole добавляет пользователю роль
func (r *UserRepository) AddRole(ctx context.Context, userID int64, roleID int) error {
	userRole := UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	return r.db.WithContext(ctx).Create(&userRole).Error
}

// RemoveRole удаляет роль у пользователя
func (r *UserRepository) RemoveRole(ctx context.Context, userID int64, roleID int) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&UserRole{}).
		Error
}

// GetRoles получает все роли пользователя
func (r *UserRepository) GetRoles(ctx context.Context, userID int64) ([]models.RoleEntity, error) {
	var roles []RoleEntity
	err := r.db.WithContext(ctx).
		Table("roles").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).
		Error
	if err != nil {
		return nil, err
	}
	var roleModels []models.RoleEntity
	for _, role := range roles {
		roleModels = append(roleModels, *ExtractRole(&role))
	}
	return roleModels, nil
}

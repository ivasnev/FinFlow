package user

import (
	"context"
	"errors"

	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"github.com/ivasnev/FinFlow/ff-id/internal/repository"
	"gorm.io/gorm"
)

// UserRepository реализует интерфейс repository.User для работы с пользователями в PostgreSQL через GORM
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
	dbUser := LoadUser(user)
	return r.db.WithContext(ctx).Create(dbUser).Error
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

// GetByIDs находит пользователей по их ID
func (r *UserRepository) GetByIDs(ctx context.Context, ids []int64) ([]*models.User, error) {
	var dbUsers []User
	err := r.db.WithContext(ctx).Where("id IN (?)", ids).Find(&dbUsers).Error
	if err != nil {
		return nil, err
	}

	users := make([]*models.User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = ExtractUser(&dbUser)
	}
	return users, nil
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
	dbUser := LoadUser(user)
	return r.db.WithContext(ctx).Save(dbUser).Error
}

// Delete удаляет пользователя
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

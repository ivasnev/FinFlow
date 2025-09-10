package repository

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uint) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
	CreateSession(ctx context.Context, session *models.UserSession) error
	FindSessionByToken(ctx context.Context, token string) (*models.UserSession, error)
	DeleteSession(ctx context.Context, token string) error
	CreateVerificationCode(ctx context.Context, code *models.VerificationCode) error
	FindVerificationCode(ctx context.Context, userID uint, codeType string) (*models.VerificationCode, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

func (r *userRepository) CreateSession(ctx context.Context, session *models.UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *userRepository) FindSessionByToken(ctx context.Context, token string) (*models.UserSession, error) {
	var session models.UserSession
	err := r.db.WithContext(ctx).Where("refresh_token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *userRepository) DeleteSession(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("refresh_token = ?", token).Delete(&models.UserSession{}).Error
}

func (r *userRepository) CreateVerificationCode(ctx context.Context, code *models.VerificationCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

func (r *userRepository) FindVerificationCode(ctx context.Context, userID uint, codeType string) (*models.VerificationCode, error) {
	var code models.VerificationCode
	err := r.db.WithContext(ctx).Where("user_id = ? AND type = ?", userID, codeType).First(&code).Error
	if err != nil {
		return nil, err
	}
	return &code, nil
} 
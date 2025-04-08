package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"gorm.io/gorm"
)

// SessionRepository реализует интерфейс для работы с сессиями в PostgreSQL через GORM
type SessionRepository struct {
	db *gorm.DB
}

// NewSessionRepository создает новый репозиторий сессий
func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

// Create создает новую сессию
func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByID находит сессию по ID
func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("сессия не найдена")
		}
		return nil, err
	}
	return &session, nil
}

// GetByRefreshToken находит сессию по refresh-токену
func (r *SessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	var session models.Session
	err := r.db.WithContext(ctx).Where("refresh_token = ?", refreshToken).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("сессия не найдена")
		}
		return nil, err
	}
	return &session, nil
}

// GetAllByUserID получает все сессии пользователя
func (r *SessionRepository) GetAllByUserID(ctx context.Context, userID int64) ([]models.Session, error) {
	var sessions []models.Session
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&sessions).Error
	return sessions, err
}

// Delete удаляет сессию
func (r *SessionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Session{}).Error
}

// DeleteAllByUserID удаляет все сессии пользователя
func (r *SessionRepository) DeleteAllByUserID(ctx context.Context, userID int64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.Session{}).Error
}

// DeleteExpired удаляет все истекшие сессии
func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.Session{}).Error
}

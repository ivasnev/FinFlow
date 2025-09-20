package postgres

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"gorm.io/gorm"
)

// LoginHistoryRepository реализует интерфейс для работы с историей входов в PostgreSQL через GORM
type LoginHistoryRepository struct {
	db *gorm.DB
}

// NewLoginHistoryRepository создает новый репозиторий истории входов
func NewLoginHistoryRepository(db *gorm.DB) *LoginHistoryRepository {
	return &LoginHistoryRepository{
		db: db,
	}
}

// Create создает новую запись в истории входов
func (r *LoginHistoryRepository) Create(ctx context.Context, history *models.LoginHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// GetAllByUserID получает всю историю входов пользователя с пагинацией
func (r *LoginHistoryRepository) GetAllByUserID(ctx context.Context, userID int64, limit, offset int) ([]models.LoginHistory, error) {
	var history []models.LoginHistory
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&history).
		Error
	return history, err
}

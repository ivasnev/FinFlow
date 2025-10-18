package login_history

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository"
	"gorm.io/gorm"
)

// LoginHistoryRepository реализует интерфейс для работы с историей входов в PostgreSQL через GORM
type LoginHistoryRepository struct {
	db *gorm.DB
}

// NewLoginHistoryRepository создает новый репозиторий истории входов
func NewLoginHistoryRepository(db *gorm.DB) repository.LoginHistory {
	return &LoginHistoryRepository{
		db: db,
	}
}

// Create создает новую запись в истории входов
func (r *LoginHistoryRepository) Create(ctx context.Context, history *models.LoginHistory) error {
	dbHistory := loadLoginHistory(history)
	return r.db.WithContext(ctx).Create(dbHistory).Error
}

// GetAllByUserID получает всю историю входов пользователя с пагинацией
func (r *LoginHistoryRepository) GetAllByUserID(ctx context.Context, userID int64, limit, offset int) ([]models.LoginHistory, error) {
	var history []LoginHistory
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&history).
		Error
	if err != nil {
		return nil, err
	}
	var historyModels []models.LoginHistory
	for _, h := range history {
		historyModels = append(historyModels, *ExtractLoginHistory(&h))
	}
	return historyModels, nil
}

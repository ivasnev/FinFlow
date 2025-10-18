package service

import (
	"context"
	"time"
)

// LoginHistoryDTO представляет данные о входе пользователя
type LoginHistoryParams struct {
	Id        int
	IpAddress string
	UserAgent *string
	CreatedAt time.Time
}

// LoginHistory определяет методы для работы с историей входов
type LoginHistory interface {
	// GetUserLoginHistory получает историю входов пользователя
	GetUserLoginHistory(ctx context.Context, userID int64, limit, offset int) ([]LoginHistoryParams, error)
}

package repository

import (
	"context"
	"time"

	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// Device определяет методы для работы с устройствами
type Device interface {
	// Create создает новое устройство
	Create(ctx context.Context, device *models.Device) error

	// GetByDeviceID находит устройство по deviceID
	GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error)

	// GetAllByUserID получает все устройства пользователя
	GetAllByUserID(ctx context.Context, userID int64) ([]models.Device, error)

	// Update обновляет данные устройства
	Update(ctx context.Context, device *models.Device) error

	// UpdateLastLogin обновляет время последнего входа
	UpdateLastLogin(ctx context.Context, deviceID string, lastLogin time.Time) error

	// Delete удаляет устройство
	Delete(ctx context.Context, id int) error
}

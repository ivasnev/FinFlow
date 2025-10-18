package service

import (
	"context"
	"time"

	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// DeviceParams представляет данные об устройстве пользователя
type DeviceParams struct {
	Id        int
	DeviceID  string
	UserAgent string
	LastLogin time.Time
}

// Device определяет методы для работы с устройствами
type Device interface {
	// GetUserDevices получает все устройства пользователя
	GetUserDevices(ctx context.Context, userID int64) ([]DeviceParams, error)

	// RemoveDevice удаляет устройство
	RemoveDevice(ctx context.Context, deviceID int, userID int64) error

	// GetOrCreateDevice получает или создает устройство по deviceID
	GetOrCreateDevice(ctx context.Context, deviceID string, userAgent string, userID int64) (*models.Device, error)
}

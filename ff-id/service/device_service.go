package service

import (
	"context"
	"time"

	"github.com/ivasnev/FinFlow/ff-id/interfaces"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// DeviceService реализует интерфейс DeviceService
type DeviceService struct {
	deviceRepository interfaces.DeviceRepository
}

// NewDeviceService создает новый сервис устройств
func NewDeviceService(deviceRepository interfaces.DeviceRepository) *DeviceService {
	return &DeviceService{
		deviceRepository: deviceRepository,
	}
}

// GetOrCreate получает или создает устройство
func (s *DeviceService) GetOrCreate(ctx context.Context, userID int64, deviceInfo *models.DeviceInfo) (*models.Device, error) {
	// Заглушка
	return &models.Device{}, nil
}

// UpdateLastLogin обновляет время последнего входа с устройства
func (s *DeviceService) UpdateLastLogin(ctx context.Context, deviceID string) error {
	// Заглушка
	return s.deviceRepository.UpdateLastLogin(ctx, deviceID, time.Now())
}

// GetAllByUserID получает все устройства пользователя
func (s *DeviceService) GetAllByUserID(ctx context.Context, userID int64) ([]models.Device, error) {
	// Заглушка
	return s.deviceRepository.GetAllByUserID(ctx, userID)
}

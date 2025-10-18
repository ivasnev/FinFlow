package device

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
)

// DeviceService реализует интерфейс для работы с устройствами
type DeviceService struct {
	deviceRepository repository.Device
}

// NewDeviceService создает новый сервис устройств
func NewDeviceService(
	deviceRepository repository.Device,
) *DeviceService {
	return &DeviceService{
		deviceRepository: deviceRepository,
	}
}

// GetUserDevices получает все устройства пользователя
func (s *DeviceService) GetUserDevices(ctx context.Context, userID int64) ([]service.DeviceParams, error) {
	devices, err := s.deviceRepository.GetAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения устройств: %w", err)
	}

	// Преобразуем в параметры устройств
	result := make([]service.DeviceParams, len(devices))
	for i, device := range devices {
		result[i] = service.DeviceParams{
			Id:        device.ID,
			DeviceID:  device.DeviceID,
			UserAgent: device.UserAgent,
			LastLogin: device.LastLogin,
		}
	}

	return result, nil
}

// RemoveDevice удаляет устройство
func (s *DeviceService) RemoveDevice(ctx context.Context, deviceID int, userID int64) error {
	// Получаем все устройства пользователя
	devices, err := s.deviceRepository.GetAllByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("ошибка получения устройств: %w", err)
	}

	// Проверяем, принадлежит ли устройство пользователю
	found := false
	for _, device := range devices {
		if device.ID == deviceID {
			found = true
			break
		}
	}

	if !found {
		return errors.New("устройство не найдено или вы не имеете прав на его удаление")
	}

	// Удаляем устройство
	return s.deviceRepository.Delete(ctx, deviceID)
}

// GetOrCreateDevice получает или создает устройство по deviceID
func (s *DeviceService) GetOrCreateDevice(ctx context.Context, deviceID string, userAgent string, userID int64) (*models.Device, error) {
	// Пытаемся найти устройство
	device, err := s.deviceRepository.GetByDeviceID(ctx, deviceID)
	if err == nil {
		// Устройство найдено, обновляем время последнего входа
		currentTime := time.Now()
		err = s.deviceRepository.UpdateLastLogin(ctx, deviceID, currentTime)
		if err != nil {
			return nil, fmt.Errorf("ошибка обновления времени последнего входа: %w", err)
		}

		// Обновляем устройство в памяти тоже
		device.LastLogin = currentTime
		return device, nil
	}

	// Устройство не найдено, создаем новое
	newDevice := &models.Device{
		UserID:    userID,
		DeviceID:  deviceID,
		UserAgent: userAgent,
		LastLogin: time.Now(),
	}

	if err := s.deviceRepository.Create(ctx, newDevice); err != nil {
		return nil, fmt.Errorf("ошибка создания устройства: %w", err)
	}

	return newDevice, nil
}

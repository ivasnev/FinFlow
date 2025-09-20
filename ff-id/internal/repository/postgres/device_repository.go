package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/ivasnev/FinFlow/ff-id/internal/models"
	"gorm.io/gorm"
)

// DeviceRepository реализует интерфейс для работы с устройствами в PostgreSQL через GORM
type DeviceRepository struct {
	db *gorm.DB
}

// NewDeviceRepository создает новый репозиторий устройств
func NewDeviceRepository(db *gorm.DB) *DeviceRepository {
	return &DeviceRepository{
		db: db,
	}
}

// Create создает новое устройство
func (r *DeviceRepository) Create(ctx context.Context, device *models.Device) error {
	return r.db.WithContext(ctx).Create(device).Error
}

// GetByDeviceID находит устройство по deviceID
func (r *DeviceRepository) GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error) {
	var device models.Device
	err := r.db.WithContext(ctx).Where("device_id = ?", deviceID).First(&device).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("устройство не найдено")
		}
		return nil, err
	}
	return &device, nil
}

// GetAllByUserID получает все устройства пользователя
func (r *DeviceRepository) GetAllByUserID(ctx context.Context, userID int64) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&devices).Error
	return devices, err
}

// Update обновляет данные устройства
func (r *DeviceRepository) Update(ctx context.Context, device *models.Device) error {
	return r.db.WithContext(ctx).Save(device).Error
}

// UpdateLastLogin обновляет время последнего входа
func (r *DeviceRepository) UpdateLastLogin(ctx context.Context, deviceID string, lastLogin time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.Device{}).
		Where("device_id = ?", deviceID).
		Update("last_login", lastLogin).
		Error
}

// Delete удаляет устройство
func (r *DeviceRepository) Delete(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Device{}, id).Error
}

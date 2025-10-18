package device

import (
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
)

// extract преобразует модель устройства базы данных в обычную модель
func extract(dbDevice *Device) *models.Device {
	if dbDevice == nil {
		return nil
	}

	return &models.Device{
		ID:        dbDevice.ID,
		UserID:    dbDevice.UserID,
		DeviceID:  dbDevice.DeviceID,
		UserAgent: dbDevice.UserAgent,
		LastLogin: dbDevice.LastLogin,
	}
}

// load преобразует обычную модель устройства в модель базы данных
func load(device *models.Device) *Device {
	if device == nil {
		return nil
	}

	return &Device{
		ID:        device.ID,
		UserID:    device.UserID,
		DeviceID:  device.DeviceID,
		UserAgent: device.UserAgent,
		LastLogin: device.LastLogin,
	}
}

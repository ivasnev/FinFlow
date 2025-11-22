package device

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/mock"
	"github.com/stretchr/testify/assert"
)

func TestDeviceService_GetUserDevices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockDevice(ctrl)
	deviceService := NewDeviceService(mockRepo)

	ctx := context.Background()
	userID := int64(1)

	// Подготовка тестовых данных
	devices := []models.Device{
		{
			ID:        1,
			UserID:    userID,
			DeviceID:  "device1",
			UserAgent: "Mozilla/5.0",
			LastLogin: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        2,
			UserID:    userID,
			DeviceID:  "device2",
			UserAgent: "Chrome/91.0",
			LastLogin: time.Now().Add(-2 * time.Hour),
		},
	}

	t.Run("успешное получение устройств", func(t *testing.T) {
		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID).
			Return(devices, nil).
			Times(1)

		result, err := deviceService.GetUserDevices(ctx, userID)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, 1, result[0].Id)
		assert.Equal(t, "device1", result[0].DeviceID)
		assert.Equal(t, "Mozilla/5.0", result[0].UserAgent)
	})

	t.Run("ошибка репозитория", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := deviceService.GetUserDevices(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestDeviceService_RemoveDevice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockDevice(ctrl)
	deviceService := NewDeviceService(mockRepo)

	ctx := context.Background()
	userID := int64(1)
	deviceID := 1

	t.Run("успешное удаление устройства", func(t *testing.T) {
		devices := []models.Device{
			{
				ID:        deviceID,
				UserID:    userID,
				DeviceID:  "device1",
				UserAgent: "Mozilla/5.0",
				LastLogin: time.Now(),
			},
		}

		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID).
			Return(devices, nil).
			Times(1)

		mockRepo.EXPECT().
			Delete(ctx, deviceID).
			Return(nil).
			Times(1)

		err := deviceService.RemoveDevice(ctx, deviceID, userID)

		assert.NoError(t, err)
	})

	t.Run("устройство не найдено", func(t *testing.T) {
		devices := []models.Device{
			{
				ID:        2, // Другое устройство
				UserID:    userID,
				DeviceID:  "device2",
				UserAgent: "Chrome/91.0",
				LastLogin: time.Now(),
			},
		}

		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID).
			Return(devices, nil).
			Times(1)

		err := deviceService.RemoveDevice(ctx, deviceID, userID)

		assert.Error(t, err)
		assert.Equal(t, "устройство не найдено или вы не имеете прав на его удаление", err.Error())
	})

	t.Run("ошибка получения устройств", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID).
			Return(nil, expectedErr).
			Times(1)

		err := deviceService.RemoveDevice(ctx, deviceID, userID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestDeviceService_GetOrCreateDevice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockDevice(ctrl)
	deviceService := NewDeviceService(mockRepo)

	ctx := context.Background()
	userID := int64(1)
	deviceID := "test-device"
	userAgent := "Mozilla/5.0"

	t.Run("устройство найдено, обновляем время входа", func(t *testing.T) {
		existingDevice := &models.Device{
			ID:        1,
			UserID:    userID,
			DeviceID:  deviceID,
			UserAgent: userAgent,
			LastLogin: time.Now().Add(-1 * time.Hour),
		}

		mockRepo.EXPECT().
			GetByDeviceID(ctx, deviceID).
			Return(existingDevice, nil).
			Times(1)

		mockRepo.EXPECT().
			UpdateLastLogin(ctx, deviceID, gomock.Any()).
			Return(nil).
			Times(1)

		result, err := deviceService.GetOrCreateDevice(ctx, deviceID, userAgent, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, deviceID, result.DeviceID)
		assert.True(t, result.LastLogin.After(time.Now().Add(-1*time.Minute)))
	})

	t.Run("устройство не найдено, создаем новое", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByDeviceID(ctx, deviceID).
			Return(nil, errors.New("device not found")).
			Times(1)

		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, device *models.Device) error {
				assert.Equal(t, userID, device.UserID)
				assert.Equal(t, deviceID, device.DeviceID)
				assert.Equal(t, userAgent, device.UserAgent)
				device.ID = 1 // Устанавливаем ID для возврата
				return nil
			}).
			Times(1)

		result, err := deviceService.GetOrCreateDevice(ctx, deviceID, userAgent, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, deviceID, result.DeviceID)
	})

	t.Run("ошибка при обновлении времени входа", func(t *testing.T) {
		existingDevice := &models.Device{
			ID:        1,
			UserID:    userID,
			DeviceID:  deviceID,
			UserAgent: userAgent,
			LastLogin: time.Now().Add(-1 * time.Hour),
		}

		mockRepo.EXPECT().
			GetByDeviceID(ctx, deviceID).
			Return(existingDevice, nil).
			Times(1)

		expectedErr := errors.New("update error")
		mockRepo.EXPECT().
			UpdateLastLogin(ctx, deviceID, gomock.Any()).
			Return(expectedErr).
			Times(1)

		result, err := deviceService.GetOrCreateDevice(ctx, deviceID, userAgent, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("ошибка при создании устройства", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByDeviceID(ctx, deviceID).
			Return(nil, errors.New("device not found")).
			Times(1)

		expectedErr := errors.New("create error")
		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(expectedErr).
			Times(1)

		result, err := deviceService.GetOrCreateDevice(ctx, deviceID, userAgent, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

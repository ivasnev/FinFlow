package device

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/mock"
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

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if len(result) != 2 {
			t.Fatalf("Ожидалось 2 устройства, получено %d", len(result))
		}

		// Проверяем первое устройство
		if result[0].Id != 1 {
			t.Errorf("Ожидался ID 1, получен %d", result[0].Id)
		}
		if result[0].DeviceID != "device1" {
			t.Errorf("Ожидался DeviceID 'device1', получен '%s'", result[0].DeviceID)
		}
		if result[0].UserAgent != "Mozilla/5.0" {
			t.Errorf("Ожидался UserAgent 'Mozilla/5.0', получен '%s'", result[0].UserAgent)
		}
	})

	t.Run("ошибка репозитория", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := deviceService.GetUserDevices(ctx, userID)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if result != nil {
			t.Fatal("Ожидался nil результат при ошибке")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
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

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}
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

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		expectedErrMsg := "устройство не найдено или вы не имеете прав на его удаление"
		if err.Error() != expectedErrMsg {
			t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedErrMsg, err.Error())
		}
	})

	t.Run("ошибка получения устройств", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetAllByUserID(ctx, userID).
			Return(nil, expectedErr).
			Times(1)

		err := deviceService.RemoveDevice(ctx, deviceID, userID)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
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

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if result == nil {
			t.Fatal("Ожидалось устройство, получен nil")
		}

		if result.DeviceID != deviceID {
			t.Errorf("Ожидался DeviceID '%s', получен '%s'", deviceID, result.DeviceID)
		}

		// Проверяем, что время обновилось
		if result.LastLogin.Before(time.Now().Add(-1 * time.Minute)) {
			t.Error("Время последнего входа должно быть обновлено")
		}
	})

	t.Run("устройство не найдено, создаем новое", func(t *testing.T) {
		mockRepo.EXPECT().
			GetByDeviceID(ctx, deviceID).
			Return(nil, errors.New("device not found")).
			Times(1)

		mockRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, device *models.Device) error {
				// Проверяем, что устройство создается с правильными данными
				if device.UserID != userID {
					t.Errorf("Ожидался UserID %d, получен %d", userID, device.UserID)
				}
				if device.DeviceID != deviceID {
					t.Errorf("Ожидался DeviceID '%s', получен '%s'", deviceID, device.DeviceID)
				}
				if device.UserAgent != userAgent {
					t.Errorf("Ожидался UserAgent '%s', получен '%s'", userAgent, device.UserAgent)
				}
				device.ID = 1 // Устанавливаем ID для возврата
				return nil
			}).
			Times(1)

		result, err := deviceService.GetOrCreateDevice(ctx, deviceID, userAgent, userID)

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if result == nil {
			t.Fatal("Ожидалось устройство, получен nil")
		}

		if result.DeviceID != deviceID {
			t.Errorf("Ожидался DeviceID '%s', получен '%s'", deviceID, result.DeviceID)
		}
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

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if result != nil {
			t.Fatal("Ожидался nil результат при ошибке")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
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

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if result != nil {
			t.Fatal("Ожидался nil результат при ошибке")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
	})
}

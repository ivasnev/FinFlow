package auth

import (
	"crypto/ed25519"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/mock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	servicemock "github.com/ivasnev/FinFlow/ff-auth/internal/service/mock"
	"github.com/stretchr/testify/assert"
)

// createMockIDAdapter создает мок адаптера для тестов
func createMockIDAdapter() *ffid.Adapter {
	// Создаем реальный адаптер с моковым URL для тестов
	// В реальных тестах можно использовать httptest сервер
	adapter, _ := ffid.NewAdapter("http://localhost:9999", nil)
	return adapter
}

func TestAuthService_ValidateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockRoleRepo := mock.NewMockRole(ctrl)
	mockSessionRepo := mock.NewMockSession(ctrl)
	mockDeviceService := servicemock.NewMockDevice(ctrl)
	mockLoginHistoryRepo := mock.NewMockLoginHistory(ctrl)
	mockTokenManager := servicemock.NewMockTokenManager(ctrl)
	mockIDClient := createMockIDAdapter()

	cfg := &config.Config{}

	authService := NewAuthService(
		cfg,
		mockUserRepo,
		mockRoleRepo,
		mockSessionRepo,
		mockDeviceService,
		mockLoginHistoryRepo,
		mockTokenManager,
		mockIDClient,
	)

	t.Run("валидный токен", func(t *testing.T) {
		token := "valid-token"
		userID := int64(1)
		roles := []string{"user", "admin"}

		mockTokenManager.EXPECT().
			ValidateToken(token).
			Return(&service.TokenPayload{
				UserID: userID,
				Roles:  roles,
				Exp:    time.Now().Add(time.Hour).Unix(),
			}, nil).
			Times(1)

		resultUserID, resultRoles, err := authService.ValidateToken(token)

		assert.NoError(t, err)
		assert.Equal(t, userID, resultUserID)
		assert.Equal(t, len(roles), len(resultRoles))
	})

	t.Run("невалидный токен", func(t *testing.T) {
		token := "invalid-token"

		mockTokenManager.EXPECT().
			ValidateToken(token).
			Return(nil, errors.New("invalid token")).
			Times(1)

		resultUserID, resultRoles, err := authService.ValidateToken(token)

		assert.Error(t, err)
		assert.Equal(t, int64(0), resultUserID)
		assert.Nil(t, resultRoles)
	})
}

func TestAuthService_GetPublicKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockRoleRepo := mock.NewMockRole(ctrl)
	mockSessionRepo := mock.NewMockSession(ctrl)
	mockDeviceService := servicemock.NewMockDevice(ctrl)
	mockLoginHistoryRepo := mock.NewMockLoginHistory(ctrl)
	mockTokenManager := servicemock.NewMockTokenManager(ctrl)
	mockIDClient := createMockIDAdapter()

	cfg := &config.Config{}

	authService := NewAuthService(
		cfg,
		mockUserRepo,
		mockRoleRepo,
		mockSessionRepo,
		mockDeviceService,
		mockLoginHistoryRepo,
		mockTokenManager,
		mockIDClient,
	)

	t.Run("получение публичного ключа", func(t *testing.T) {
		expectedKey := []byte("test-public-key")

		mockTokenManager.EXPECT().
			GetPublicKey().
			Return(ed25519.PublicKey(expectedKey)).
			Times(1)

		result := authService.GetPublicKey()

		assert.NotNil(t, result)
		assert.Equal(t, len(expectedKey), len(result))
	})
}

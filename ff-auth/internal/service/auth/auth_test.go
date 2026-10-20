package auth

import (
	"context"
	"crypto/ed25519"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/mock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	servicemock "github.com/ivasnev/FinFlow/ff-auth/internal/service/mock"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// createMockIDAdapter создаёт независимый источник идентификаторов для unit-тестов.
func createMockIDAdapter(ctrl *gomock.Controller) *servicemock.MockIdentity {
	return servicemock.NewMockIdentity(ctrl)
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
	mockIDClient := createMockIDAdapter(ctrl)

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
	mockIDClient := createMockIDAdapter(ctrl)

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

func TestAuthService_GenerateTokenPair(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockRoleRepo := mock.NewMockRole(ctrl)
	mockSessionRepo := mock.NewMockSession(ctrl)
	mockDeviceService := servicemock.NewMockDevice(ctrl)
	mockLoginHistoryRepo := mock.NewMockLoginHistory(ctrl)
	mockTokenManager := servicemock.NewMockTokenManager(ctrl)
	mockIDClient := createMockIDAdapter(ctrl)

	cfg := &config.Config{}
	cfg.Auth.AccessTokenDuration = 15
	cfg.Auth.RefreshTokenDuration = 10080

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

	ctx := context.Background()
	userID := int64(1)
	roles := []string{"user", "admin"}

	t.Run("успешная генерация пары токенов", func(t *testing.T) {
		accessToken := "access-token"
		refreshToken := "refresh-token"
		expiresAt := time.Now().Add(15 * time.Minute).Unix()

		mockTokenManager.EXPECT().
			GenerateTokenPair(userID, roles, 15*time.Minute, 10080*time.Minute).
			Return(accessToken, refreshToken, expiresAt, expiresAt+7*24*3600, nil).
			Times(1)

		resultAccess, resultRefresh, resultExpiresAt, err := authService.GenerateTokenPair(ctx, userID, roles)

		assert.NoError(t, err)
		assert.Equal(t, accessToken, resultAccess)
		assert.Equal(t, refreshToken, resultRefresh)
		assert.Equal(t, expiresAt, resultExpiresAt)
	})

	t.Run("ошибка генерации токенов", func(t *testing.T) {
		expectedErr := errors.New("token generation error")

		mockTokenManager.EXPECT().
			GenerateTokenPair(userID, roles, 15*time.Minute, 10080*time.Minute).
			Return("", "", int64(0), int64(0), expectedErr).
			Times(1)

		resultAccess, resultRefresh, resultExpiresAt, err := authService.GenerateTokenPair(ctx, userID, roles)

		assert.Error(t, err)
		assert.Empty(t, resultAccess)
		assert.Empty(t, resultRefresh)
		assert.Equal(t, int64(0), resultExpiresAt)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestAuthService_RecordLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockRoleRepo := mock.NewMockRole(ctrl)
	mockSessionRepo := mock.NewMockSession(ctrl)
	mockDeviceService := servicemock.NewMockDevice(ctrl)
	mockLoginHistoryRepo := mock.NewMockLoginHistory(ctrl)
	mockTokenManager := servicemock.NewMockTokenManager(ctrl)
	mockIDClient := createMockIDAdapter(ctrl)

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

	ctx := context.Background()
	userID := int64(1)
	ipAddress := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	t.Run("успешная запись истории входа", func(t *testing.T) {
		mockLoginHistoryRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, history *models.LoginHistory) error {
				assert.Equal(t, userID, history.UserID)
				assert.Equal(t, ipAddress, history.IPAddress)
				assert.Equal(t, userAgent, history.UserAgent)
				return nil
			}).
			Times(1)

		err := authService.RecordLogin(ctx, userID, ipAddress, userAgent)

		assert.NoError(t, err)
	})

	t.Run("ошибка записи истории входа", func(t *testing.T) {
		expectedErr := errors.New("database error")

		mockLoginHistoryRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(expectedErr).
			Times(1)

		err := authService.RecordLogin(ctx, userID, ipAddress, userAgent)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

// TestAuthService_Login проверяет актуальные логины, пароль и отказ без источника профиля.
func TestAuthService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockRoleRepo := mock.NewMockRole(ctrl)
	mockSessionRepo := mock.NewMockSession(ctrl)
	mockDeviceService := servicemock.NewMockDevice(ctrl)
	mockLoginHistoryRepo := mock.NewMockLoginHistory(ctrl)
	mockTokenManager := servicemock.NewMockTokenManager(ctrl)
	mockIDClient := createMockIDAdapter(ctrl)

	cfg := &config.Config{}
	cfg.Auth.AccessTokenDuration = 15
	cfg.Auth.RefreshTokenDuration = 10080

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

	ctx := context.Background()
	password := "password123"
	hash, _ := hashPassword(password, 10)
	for _, login := range []string{"fresh@example.com", "fresh"} {
		mockIDClient.EXPECT().FindIdentity(ctx, login, ffid.LookupLogin).Return(&ffid.UserDTO{ID: 1, Email: "fresh@example.com", Nickname: "fresh"}, nil)
		mockUserRepo.EXPECT().GetByID(ctx, int64(1)).Return(&models.User{ID: 1, Email: "stale@example.com", Nickname: "stale", PasswordHash: hash}, nil)
		mockUserRepo.EXPECT().GetRoles(ctx, int64(1)).Return([]models.RoleEntity{{ID: 1, Name: "user"}}, nil)
		mockTokenManager.EXPECT().GenerateTokenPair(int64(1), []string{"user"}, 15*time.Minute, 10080*time.Minute).Return("access", "refresh", time.Now().Add(time.Minute).Unix(), time.Now().Add(time.Hour).Unix(), nil)
		mockDeviceService.EXPECT().GetOrCreateDevice(gomock.Any(), gomock.Any(), gomock.Any(), int64(1)).Return(nil, nil)
		mockSessionRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		mockLoginHistoryRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)
		got, err := authService.Login(ctx, service.LoginParams{Login: login, Password: password})
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, "fresh", got.User.Nickname)
			assert.Equal(t, "fresh@example.com", got.User.Email)
		}
	}
	// Сценарий: старый логин отсутствует в источнике профиля, локальный поиск не выполняется.
	mockIDClient.EXPECT().FindIdentity(ctx, "stale", ffid.LookupLogin).Return(nil, ffid.ErrNotFound)
	_, err := authService.Login(ctx, service.LoginParams{Login: "stale", Password: password})
	assert.Error(t, err)
	// Сценарий: недоступность ff-id не разрешает вход по локальной копии.
	mockIDClient.EXPECT().FindIdentity(ctx, "fresh", ffid.LookupLogin).Return(nil, ffid.ErrUnavailable)
	_, err = authService.Login(ctx, service.LoginParams{Login: "fresh", Password: password})
	assert.ErrorIs(t, err, ffid.ErrUnavailable)
	// Сценарий: правильный актуальный логин не обходит проверку пароля.
	mockIDClient.EXPECT().FindIdentity(ctx, "fresh", ffid.LookupLogin).Return(&ffid.UserDTO{ID: 1}, nil)
	mockUserRepo.EXPECT().GetByID(ctx, int64(1)).Return(&models.User{ID: 1, PasswordHash: hash}, nil)
	_, err = authService.Login(ctx, service.LoginParams{Login: "fresh", Password: "wrong"})
	assert.Error(t, err)
}

// TestAuthService_RefreshToken проверяет обновление сессии с актуальным профилем.
func TestAuthService_RefreshToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockRoleRepo := mock.NewMockRole(ctrl)
	mockSessionRepo := mock.NewMockSession(ctrl)
	mockDeviceService := servicemock.NewMockDevice(ctrl)
	mockLoginHistoryRepo := mock.NewMockLoginHistory(ctrl)
	mockTokenManager := servicemock.NewMockTokenManager(ctrl)
	mockIDClient := createMockIDAdapter(ctrl)

	cfg := &config.Config{}
	cfg.Auth.AccessTokenDuration = 15
	cfg.Auth.RefreshTokenDuration = 10080

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

	ctx := context.Background()
	refreshToken := "old-refresh-token"
	userID := int64(1)

	t.Run("успешное обновление токена", func(t *testing.T) {
		sessionID := uuid.New()
		session := &models.Session{
			ID:           sessionID,
			UserID:       userID,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(time.Hour),
			CreatedAt:    time.Now().Add(-time.Hour),
		}

		user := &models.User{
			ID:        userID,
			Email:     "test@example.com",
			Nickname:  "testuser",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}

		roles := []models.RoleEntity{
			{ID: 1, Name: "user"},
		}

		newAccessToken := "new-access-token"
		newRefreshToken := "new-refresh-token"
		expiresAt := time.Now().Add(15 * time.Minute).Unix()

		mockSessionRepo.EXPECT().
			GetByRefreshToken(ctx, refreshToken).
			Return(session, nil).
			Times(1)

		mockUserRepo.EXPECT().
			GetByID(ctx, userID).
			Return(user, nil).
			Times(1)

		mockIDClient.EXPECT().GetIdentity(ctx, userID).Return(&ffid.UserDTO{ID: userID, Email: "test@example.com", Nickname: "testuser"}, nil)

		mockUserRepo.EXPECT().
			GetRoles(ctx, userID).
			Return(roles, nil).
			Times(1)

		mockTokenManager.EXPECT().
			GenerateTokenPair(userID, []string{"user"}, 15*time.Minute, 10080*time.Minute).
			Return(newAccessToken, newRefreshToken, expiresAt, expiresAt+7*24*3600, nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Delete(ctx, sessionID).
			Return(nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, newSession *models.Session) error {
				assert.Equal(t, userID, newSession.UserID)
				assert.Equal(t, newRefreshToken, newSession.RefreshToken)
				return nil
			}).
			Times(1)

		result, err := authService.RefreshToken(ctx, refreshToken)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newAccessToken, result.AccessToken)
		assert.Equal(t, newRefreshToken, result.RefreshToken)
		assert.Equal(t, userID, result.User.Id)
	})

	t.Run("недействительный refresh токен", func(t *testing.T) {
		mockSessionRepo.EXPECT().
			GetByRefreshToken(ctx, refreshToken).
			Return(nil, errors.New("session not found")).
			Times(1)

		result, err := authService.RefreshToken(ctx, refreshToken)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "недействительный refresh-токен", err.Error())
	})

	t.Run("истекший refresh токен", func(t *testing.T) {
		sessionID := uuid.New()
		session := &models.Session{
			ID:           sessionID,
			UserID:       userID,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(-time.Hour), // Истекший токен
			CreatedAt:    time.Now().Add(-2 * time.Hour),
		}

		mockSessionRepo.EXPECT().
			GetByRefreshToken(ctx, refreshToken).
			Return(session, nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Delete(ctx, sessionID).
			Return(nil).
			Times(1)

		result, err := authService.RefreshToken(ctx, refreshToken)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "истек срок действия refresh-токена", err.Error())
	})
}

func TestAuthService_Logout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockRoleRepo := mock.NewMockRole(ctrl)
	mockSessionRepo := mock.NewMockSession(ctrl)
	mockDeviceService := servicemock.NewMockDevice(ctrl)
	mockLoginHistoryRepo := mock.NewMockLoginHistory(ctrl)
	mockTokenManager := servicemock.NewMockTokenManager(ctrl)
	mockIDClient := createMockIDAdapter(ctrl)

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

	ctx := context.Background()
	refreshToken := "refresh-token"
	sessionID := uuid.New()

	t.Run("успешный выход", func(t *testing.T) {
		session := &models.Session{
			ID:           sessionID,
			UserID:       1,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(time.Hour),
			CreatedAt:    time.Now().Add(-time.Hour),
		}

		mockSessionRepo.EXPECT().
			GetByRefreshToken(ctx, refreshToken).
			Return(session, nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Delete(ctx, sessionID).
			Return(nil).
			Times(1)

		err := authService.Logout(ctx, refreshToken)

		assert.NoError(t, err)
	})

	t.Run("недействительный refresh токен", func(t *testing.T) {
		mockSessionRepo.EXPECT().
			GetByRefreshToken(ctx, refreshToken).
			Return(nil, errors.New("session not found")).
			Times(1)

		err := authService.Logout(ctx, refreshToken)

		assert.Error(t, err)
		assert.Equal(t, "недействительный refresh-токен", err.Error())
	})

	t.Run("ошибка удаления сессии", func(t *testing.T) {
		session := &models.Session{
			ID:           sessionID,
			UserID:       1,
			RefreshToken: refreshToken,
			ExpiresAt:    time.Now().Add(time.Hour),
			CreatedAt:    time.Now().Add(-time.Hour),
		}

		expectedErr := errors.New("delete error")

		mockSessionRepo.EXPECT().
			GetByRefreshToken(ctx, refreshToken).
			Return(session, nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Delete(ctx, sessionID).
			Return(expectedErr).
			Times(1)

		err := authService.Logout(ctx, refreshToken)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

// TestAuthService_Register проверяет уникальность идентификаторов и отказ без источника.
func TestAuthService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock.NewMockUser(ctrl)
	mockRoleRepo := mock.NewMockRole(ctrl)
	mockSessionRepo := mock.NewMockSession(ctrl)
	mockDeviceService := servicemock.NewMockDevice(ctrl)
	mockLoginHistoryRepo := mock.NewMockLoginHistory(ctrl)
	mockTokenManager := servicemock.NewMockTokenManager(ctrl)
	mockIDClient := createMockIDAdapter(ctrl)

	cfg := &config.Config{}
	cfg.Auth.PasswordHashCost = 10
	cfg.Auth.AccessTokenDuration = 15
	cfg.Auth.RefreshTokenDuration = 10080

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

	ctx := context.Background()
	params := service.RegisterParams{Email: "test@example.com", Nickname: "testuser", Password: "password123"}
	// Сценарий: занятость email и nickname определяется ff-id.
	mockIDClient.EXPECT().FindIdentity(ctx, params.Email, ffid.LookupEmail).Return(&ffid.UserDTO{ID: 1}, nil)
	_, err := authService.Register(ctx, params)
	assert.Error(t, err)
	mockIDClient.EXPECT().FindIdentity(ctx, params.Email, ffid.LookupEmail).Return(nil, ffid.ErrNotFound)
	mockIDClient.EXPECT().FindIdentity(ctx, params.Nickname, ffid.LookupNickname).Return(&ffid.UserDTO{ID: 1}, nil)
	_, err = authService.Register(ctx, params)
	assert.Error(t, err)
	// Сценарий: недоступность источника не допускает создания локальной учётной записи.
	mockIDClient.EXPECT().FindIdentity(ctx, params.Email, ffid.LookupEmail).Return(nil, ffid.ErrUnavailable)
	_, err = authService.Register(ctx, params)
	assert.ErrorIs(t, err, ffid.ErrUnavailable)
	mockIDClient.EXPECT().FindIdentity(ctx, params.Email, ffid.LookupEmail).Return(nil, ffid.ErrNotFound)
	mockIDClient.EXPECT().FindIdentity(ctx, params.Nickname, ffid.LookupNickname).Return(nil, ffid.ErrNotFound)
	expected := errors.New("create failure")
	mockUserRepo.EXPECT().Create(ctx, gomock.Any()).Return(expected)
	_, err = authService.Register(ctx, params)
	assert.ErrorIs(t, err, expected)
}

// hashPassword создаёт хеш пароля для проверки входа.
func hashPassword(password string, cost int) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

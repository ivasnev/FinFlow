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

func TestAuthService_GenerateTokenPair(t *testing.T) {
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
			Return(accessToken, refreshToken, expiresAt, nil).
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
			Return("", "", int64(0), expectedErr).
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

func TestAuthService_Login(t *testing.T) {
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
	hashedPassword, _ := hashPassword(password, 10)

	t.Run("успешный вход по email", func(t *testing.T) {
		email := "test@example.com"
		userID := int64(1)
		user := &models.User{
			ID:           userID,
			Email:        email,
			PasswordHash: hashedPassword,
			Nickname:     "testuser",
			CreatedAt:    time.Now().Add(-24 * time.Hour),
			UpdatedAt:    time.Now().Add(-1 * time.Hour),
		}

		roles := []models.RoleEntity{
			{ID: 1, Name: "user"},
		}

		accessToken := "access-token"
		refreshToken := "refresh-token"
		expiresAt := time.Now().Add(15 * time.Minute).Unix()

		mockUserRepo.EXPECT().
			GetByEmail(ctx, email).
			Return(user, nil).
			Times(1)

		mockUserRepo.EXPECT().
			GetRoles(ctx, userID).
			Return(roles, nil).
			Times(1)

		mockTokenManager.EXPECT().
			GenerateTokenPair(userID, []string{"user"}, 15*time.Minute, 10080*time.Minute).
			Return(accessToken, refreshToken, expiresAt, nil).
			Times(1)

		mockDeviceService.EXPECT().
			GetOrCreateDevice(ctx, gomock.Any(), "Mozilla/5.0", userID).
			Return(&models.Device{ID: 1, UserID: userID, DeviceID: "device1"}, nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Create(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, session *models.Session) error {
				assert.Equal(t, userID, session.UserID)
				assert.Equal(t, refreshToken, session.RefreshToken)
				return nil
			}).
			Times(1)

		mockLoginHistoryRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil).
			Times(1)

		params := service.LoginParams{
			Login:     email,
			Password:  password,
			UserAgent: "Mozilla/5.0",
			IpAddress: "192.168.1.1",
		}

		result, err := authService.Login(ctx, params)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, accessToken, result.AccessToken)
		assert.Equal(t, refreshToken, result.RefreshToken)
		assert.Equal(t, userID, result.User.Id)
		assert.Equal(t, email, result.User.Email)
	})

	t.Run("успешный вход по nickname", func(t *testing.T) {
		nickname := "testuser"
		userID := int64(1)
		user := &models.User{
			ID:           userID,
			Email:        "test@example.com",
			PasswordHash: hashedPassword,
			Nickname:     nickname,
			CreatedAt:    time.Now().Add(-24 * time.Hour),
			UpdatedAt:    time.Now().Add(-1 * time.Hour),
		}

		roles := []models.RoleEntity{
			{ID: 1, Name: "user"},
		}

		accessToken := "access-token"
		refreshToken := "refresh-token"
		expiresAt := time.Now().Add(15 * time.Minute).Unix()

		mockUserRepo.EXPECT().
			GetByNickname(ctx, nickname).
			Return(user, nil).
			Times(1)

		mockUserRepo.EXPECT().
			GetRoles(ctx, userID).
			Return(roles, nil).
			Times(1)

		mockTokenManager.EXPECT().
			GenerateTokenPair(userID, []string{"user"}, 15*time.Minute, 10080*time.Minute).
			Return(accessToken, refreshToken, expiresAt, nil).
			Times(1)

		mockDeviceService.EXPECT().
			GetOrCreateDevice(ctx, gomock.Any(), "Mozilla/5.0", userID).
			Return(&models.Device{ID: 1, UserID: userID, DeviceID: "device1"}, nil).
			Times(1)

		mockSessionRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil).
			Times(1)

		mockLoginHistoryRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(nil).
			Times(1)

		params := service.LoginParams{
			Login:     nickname,
			Password:  password,
			UserAgent: "Mozilla/5.0",
			IpAddress: "192.168.1.1",
		}

		result, err := authService.Login(ctx, params)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, nickname, result.User.Nickname)
	})

	t.Run("неверный логин", func(t *testing.T) {
		email := "nonexistent@example.com"

		mockUserRepo.EXPECT().
			GetByEmail(ctx, email).
			Return(nil, errors.New("user not found")).
			Times(1)

		params := service.LoginParams{
			Login:     email,
			Password:  password,
			UserAgent: "Mozilla/5.0",
			IpAddress: "192.168.1.1",
		}

		result, err := authService.Login(ctx, params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "неверный логин или пароль", err.Error())
	})

	t.Run("неверный пароль", func(t *testing.T) {
		email := "test@example.com"
		userID := int64(1)
		user := &models.User{
			ID:           userID,
			Email:        email,
			PasswordHash: hashedPassword,
			Nickname:     "testuser",
		}

		mockUserRepo.EXPECT().
			GetByEmail(ctx, email).
			Return(user, nil).
			Times(1)

		params := service.LoginParams{
			Login:     email,
			Password:  "wrongpassword",
			UserAgent: "Mozilla/5.0",
			IpAddress: "192.168.1.1",
		}

		result, err := authService.Login(ctx, params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "неверный логин или пароль", err.Error())
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
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

		mockUserRepo.EXPECT().
			GetRoles(ctx, userID).
			Return(roles, nil).
			Times(1)

		mockTokenManager.EXPECT().
			GenerateTokenPair(userID, []string{"user"}, 15*time.Minute, 10080*time.Minute).
			Return(newAccessToken, newRefreshToken, expiresAt, nil).
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

func TestAuthService_Register(t *testing.T) {
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

	t.Run("email уже существует", func(t *testing.T) {
		email := "existing@example.com"
		existingUser := &models.User{
			ID:    1,
			Email: email,
		}

		mockUserRepo.EXPECT().
			GetByEmail(ctx, email).
			Return(existingUser, nil).
			Times(1)

		params := service.RegisterParams{
			Email:    email,
			Password: "password123",
			Nickname: "testuser",
		}

		result, err := authService.Register(ctx, params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "пользователь с таким email уже существует", err.Error())
	})

	t.Run("никнейм уже существует", func(t *testing.T) {
		email := "test@example.com"
		nickname := "existinguser"

		mockUserRepo.EXPECT().
			GetByEmail(ctx, email).
			Return(nil, errors.New("not found")).
			Times(1)

		existingUser := &models.User{
			ID:       1,
			Nickname: nickname,
		}

		mockUserRepo.EXPECT().
			GetByNickname(ctx, nickname).
			Return(existingUser, nil).
			Times(1)

		params := service.RegisterParams{
			Email:    email,
			Password: "password123",
			Nickname: nickname,
		}

		result, err := authService.Register(ctx, params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "пользователь с таким никнеймом уже существует", err.Error())
	})

	t.Run("ошибка создания пользователя", func(t *testing.T) {
		email := "test@example.com"
		nickname := "testuser"

		mockUserRepo.EXPECT().
			GetByEmail(ctx, email).
			Return(nil, errors.New("not found")).
			Times(1)

		mockUserRepo.EXPECT().
			GetByNickname(ctx, nickname).
			Return(nil, errors.New("not found")).
			Times(1)

		expectedErr := errors.New("create error")

		mockUserRepo.EXPECT().
			Create(ctx, gomock.Any()).
			Return(expectedErr).
			Times(1)

		params := service.RegisterParams{
			Email:    email,
			Password: "password123",
			Nickname: nickname,
		}

		result, err := authService.Register(ctx, params)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "ошибка создания пользователя")
	})

}

// hashPassword хеширует пароль для тестов
func hashPassword(password string, cost int) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

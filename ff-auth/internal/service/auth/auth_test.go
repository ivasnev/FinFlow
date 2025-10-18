package auth

// import (
// 	"context"
// 	"crypto/ed25519"
// 	"errors"
// 	"testing"
// 	"time"

// 	"github.com/golang/mock/gomock"
// 	"github.com/google/uuid"
// 	"github.com/ivasnev/FinFlow/ff-auth/internal/common/config"
// 	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
// 	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/mock"
// 	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
// 	servicemock "github.com/ivasnev/FinFlow/ff-auth/internal/service/mock"
// 	idclient "github.com/ivasnev/FinFlow/ff-id/pkg/client"
// 	"golang.org/x/crypto/bcrypt"
// )

// func TestAuthService_Register(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockUserRepo := mock.NewMockUser(ctrl)
// 	mockRoleRepo := mock.NewMockRole(ctrl)
// 	mockSessionRepo := mock.NewMockSession(ctrl)
// 	mockDeviceService := servicemock.NewMockDevice(ctrl)
// 	mockLoginHistoryRepo := mock.NewMockLoginHistory(ctrl)
// 	mockTokenManager := servicemock.NewMockTokenManager(ctrl)
// 	mockIDClient := NewMockIDClient() // Мок для ID клиента

// 	config := &config.Config{}
// 	config.Auth.PasswordHashCost = 10

// 	authService := NewAuthService(
// 		config,
// 		mockUserRepo,
// 		mockRoleRepo,
// 		mockSessionRepo,
// 		mockDeviceService,
// 		mockLoginHistoryRepo,
// 		mockTokenManager,
// 		mockIDClient,
// 	)

// 	ctx := context.Background()

// 	t.Run("успешная регистрация", func(t *testing.T) {
// 		params := service.RegisterParams{
// 			Email:    "test@example.com",
// 			Password: "password123",
// 			Nickname: "testuser",
// 		}

// 		// Проверка, что пользователь с таким email не существует
// 		mockUserRepo.EXPECT().
// 			GetByEmail(ctx, params.Email).
// 			Return(nil, errors.New("not found")).
// 			Times(1)

// 		// Проверка, что пользователь с таким никнеймом не существует
// 		mockUserRepo.EXPECT().
// 			GetByNickname(ctx, params.Nickname).
// 			Return(nil, errors.New("not found")).
// 			Times(1)

// 		// Создание пользователя
// 		mockUserRepo.EXPECT().
// 			Create(ctx, gomock.Any()).
// 			DoAndReturn(func(ctx context.Context, user *models.User) error {
// 				user.ID = 1 // Устанавливаем ID для возврата
// 				return nil
// 			}).
// 			Times(1)

// 		// Настройка мока ID клиента для успешной регистрации
// 		mockIDClient.SetRegisterUserFunc(func(ctx context.Context, req *idclient.RegisterUserRequest) (*MockUserDTO, error) {
// 			return &MockUserDTO{
// 				ID:       req.UserID,
// 				Email:    req.Email,
// 				Nickname: req.Nickname,
// 			}, nil
// 		})

// 		// Получение роли "user"
// 		role := &models.RoleEntity{ID: 1, Name: "user"}
// 		mockRoleRepo.EXPECT().
// 			GetByName(ctx, "user").
// 			Return(role, nil).
// 			Times(1)

// 		// Назначение роли пользователю
// 		mockUserRepo.EXPECT().
// 			AddRole(ctx, int64(1), 1).
// 			Return(nil).
// 			Times(1)

// 		// Генерация токенов
// 		mockTokenManager.EXPECT().
// 			GenerateTokenPair(int64(1), []string{"user"}, gomock.Any(), gomock.Any()).
// 			Return("access-token", "refresh-token", int64(1234567890), nil).
// 			Times(1)

// 		// Создание сессии
// 		mockSessionRepo.EXPECT().
// 			Create(ctx, gomock.Any()).
// 			DoAndReturn(func(ctx context.Context, session *models.Session) error {
// 				session.ID = uuid.New() // Устанавливаем ID для возврата
// 				return nil
// 			}).
// 			Times(1)

// 		// Создание устройства
// 		device := &models.Device{
// 			ID:        1,
// 			UserID:    1,
// 			DeviceID:  "test-device",
// 			UserAgent: "Mozilla/5.0",
// 			LastLogin: time.Now(),
// 		}
// 		mockDeviceService.EXPECT().
// 			GetOrCreateDevice(ctx, gomock.Any(), gomock.Any(), int64(1)).
// 			Return(device, nil).
// 			Times(1)

// 		// Запись в историю входов
// 		mockLoginHistoryRepo.EXPECT().
// 			Create(ctx, gomock.Any()).
// 			Return(nil).
// 			Times(1)

// 		result, err := authService.Register(ctx, params)

// 		if err != nil {
// 			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
// 		}

// 		if result == nil {
// 			t.Fatal("Ожидались данные доступа, получен nil")
// 		}

// 		if result.AccessToken != "access-token" {
// 			t.Errorf("Ожидался access token 'access-token', получен '%s'", result.AccessToken)
// 		}

// 		if result.RefreshToken != "refresh-token" {
// 			t.Errorf("Ожидался refresh token 'refresh-token', получен '%s'", result.RefreshToken)
// 		}

// 		if result.User.Id != 1 {
// 			t.Errorf("Ожидался UserID 1, получен %d", result.User.Id)
// 		}
// 	})

// 	t.Run("пользователь с таким email уже существует", func(t *testing.T) {
// 		params := service.RegisterParams{
// 			Email:    "existing@example.com",
// 			Password: "password123",
// 			Nickname: "newuser",
// 		}

// 		existingUser := &models.User{
// 			ID:       1,
// 			Email:    "existing@example.com",
// 			Nickname: "existinguser",
// 		}

// 		mockUserRepo.EXPECT().
// 			GetByEmail(ctx, params.Email).
// 			Return(existingUser, nil).
// 			Times(1)

// 		result, err := authService.Register(ctx, params)

// 		if err == nil {
// 			t.Fatal("Ожидалась ошибка, получен успех")
// 		}

// 		if result != nil {
// 			t.Fatal("Ожидался nil результат при ошибке")
// 		}

// 		expectedErrMsg := "пользователь с таким email уже существует"
// 		if err.Error() != expectedErrMsg {
// 			t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedErrMsg, err.Error())
// 		}
// 	})

// 	t.Run("пользователь с таким никнеймом уже существует", func(t *testing.T) {
// 		params := service.RegisterParams{
// 			Email:    "new@example.com",
// 			Password: "password123",
// 			Nickname: "existinguser",
// 		}

// 		existingUser := &models.User{
// 			ID:       1,
// 			Email:    "existing@example.com",
// 			Nickname: "existinguser",
// 		}

// 		mockUserRepo.EXPECT().
// 			GetByEmail(ctx, params.Email).
// 			Return(nil, errors.New("not found")).
// 			Times(1)

// 		mockUserRepo.EXPECT().
// 			GetByNickname(ctx, params.Nickname).
// 			Return(existingUser, nil).
// 			Times(1)

// 		result, err := authService.Register(ctx, params)

// 		if err == nil {
// 			t.Fatal("Ожидалась ошибка, получен успех")
// 		}

// 		if result != nil {
// 			t.Fatal("Ожидался nil результат при ошибке")
// 		}

// 		expectedErrMsg := "пользователь с таким никнеймом уже существует"
// 		if err.Error() != expectedErrMsg {
// 			t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedErrMsg, err.Error())
// 		}
// 	})

// 	t.Run("ошибка создания пользователя", func(t *testing.T) {
// 		params := service.RegisterParams{
// 			Email:    "test@example.com",
// 			Password: "password123",
// 			Nickname: "testuser",
// 		}

// 		mockUserRepo.EXPECT().
// 			GetByEmail(ctx, params.Email).
// 			Return(nil, errors.New("not found")).
// 			Times(1)

// 		mockUserRepo.EXPECT().
// 			GetByNickname(ctx, params.Nickname).
// 			Return(nil, errors.New("not found")).
// 			Times(1)

// 		expectedErr := errors.New("database error")
// 		mockUserRepo.EXPECT().
// 			Create(ctx, gomock.Any()).
// 			Return(expectedErr).
// 			Times(1)

// 		result, err := authService.Register(ctx, params)

// 		if err == nil {
// 			t.Fatal("Ожидалась ошибка, получен успех")
// 		}

// 		if result != nil {
// 			t.Fatal("Ожидался nil результат при ошибке")
// 		}

// 		if !errors.Is(err, expectedErr) {
// 			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
// 		}
// 	})
// }

// func TestAuthService_Login(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockUserRepo := mock.NewMockUser(ctrl)
// 	mockRoleRepo := mock.NewMockRole(ctrl)
// 	mockSessionRepo := mock.NewMockSession(ctrl)
// 	mockDeviceService := servicemock.NewMockDevice(ctrl)
// 	mockLoginHistoryRepo := mock.NewMockLoginHistory(ctrl)
// 	mockTokenManager := servicemock.NewMockTokenManager(ctrl)
// 	mockIDClient := NewMockIDClient()

// 	config := &config.Config{}
// 	config.Auth.PasswordHashCost = 10

// 	authService := NewAuthService(
// 		config,
// 		mockUserRepo,
// 		mockRoleRepo,
// 		mockSessionRepo,
// 		mockDeviceService,
// 		mockLoginHistoryRepo,
// 		mockTokenManager,
// 		mockIDClient,
// 	)

// 	ctx := context.Background()

// 	t.Run("успешный вход по email", func(t *testing.T) {
// 		params := service.LoginParams{
// 			Login:     "test@example.com",
// 			Password:  "password123",
// 			UserAgent: "Mozilla/5.0",
// 			IpAddress: "192.168.1.1",
// 		}

// 		// Хешируем пароль для теста
// 		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 10)

// 		user := &models.User{
// 			ID:           1,
// 			Email:        "test@example.com",
// 			PasswordHash: string(hashedPassword),
// 			Nickname:     "testuser",
// 		}

// 		roles := []models.RoleEntity{
// 			{ID: 1, Name: "user"},
// 		}

// 		device := &models.Device{
// 			ID:        1,
// 			UserID:    1,
// 			DeviceID:  "test-device",
// 			UserAgent: "Mozilla/5.0",
// 			LastLogin: time.Now(),
// 		}

// 		// Поиск пользователя по email
// 		mockUserRepo.EXPECT().
// 			GetByEmail(ctx, params.Login).
// 			Return(user, nil).
// 			Times(1)

// 		// Получение ролей
// 		mockUserRepo.EXPECT().
// 			GetRoles(ctx, user.ID).
// 			Return(roles, nil).
// 			Times(1)

// 		// Генерация токенов
// 		mockTokenManager.EXPECT().
// 			GenerateTokenPair(user.ID, []string{"user"}, gomock.Any(), gomock.Any()).
// 			Return("access-token", "refresh-token", int64(1234567890), nil).
// 			Times(1)

// 		// Создание сессии
// 		mockSessionRepo.EXPECT().
// 			Create(ctx, gomock.Any()).
// 			DoAndReturn(func(ctx context.Context, session *models.Session) error {
// 				session.ID = uuid.New()
// 				return nil
// 			}).
// 			Times(1)

// 		// Создание/получение устройства
// 		mockDeviceService.EXPECT().
// 			GetOrCreateDevice(ctx, gomock.Any(), params.UserAgent, user.ID).
// 			Return(device, nil).
// 			Times(1)

// 		// Запись в историю входов
// 		mockLoginHistoryRepo.EXPECT().
// 			Create(ctx, gomock.Any()).
// 			Return(nil).
// 			Times(1)

// 		result, err := authService.Login(ctx, params)

// 		if err != nil {
// 			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
// 		}

// 		if result == nil {
// 			t.Fatal("Ожидались данные доступа, получен nil")
// 		}

// 		if result.AccessToken != "access-token" {
// 			t.Errorf("Ожидался access token 'access-token', получен '%s'", result.AccessToken)
// 		}

// 		if result.User.Id != 1 {
// 			t.Errorf("Ожидался UserID 1, получен %d", result.User.Id)
// 		}
// 	})

// 	t.Run("неверный пароль", func(t *testing.T) {
// 		params := service.LoginParams{
// 			Login:     "test@example.com",
// 			Password:  "wrongpassword",
// 			UserAgent: "Mozilla/5.0",
// 			IpAddress: "192.168.1.1",
// 		}

// 		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 10)

// 		user := &models.User{
// 			ID:           1,
// 			Email:        "test@example.com",
// 			PasswordHash: string(hashedPassword),
// 			Nickname:     "testuser",
// 		}

// 		mockUserRepo.EXPECT().
// 			GetByEmail(ctx, params.Login).
// 			Return(user, nil).
// 			Times(1)

// 		result, err := authService.Login(ctx, params)

// 		if err == nil {
// 			t.Fatal("Ожидалась ошибка, получен успех")
// 		}

// 		if result != nil {
// 			t.Fatal("Ожидался nil результат при ошибке")
// 		}

// 		expectedErrMsg := "неверный логин или пароль"
// 		if err.Error() != expectedErrMsg {
// 			t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedErrMsg, err.Error())
// 		}
// 	})

// 	t.Run("пользователь не найден", func(t *testing.T) {
// 		params := service.LoginParams{
// 			Login:     "nonexistent@example.com",
// 			Password:  "password123",
// 			UserAgent: "Mozilla/5.0",
// 			IpAddress: "192.168.1.1",
// 		}

// 		mockUserRepo.EXPECT().
// 			GetByEmail(ctx, params.Login).
// 			Return(nil, errors.New("not found")).
// 			Times(1)

// 		mockUserRepo.EXPECT().
// 			GetByNickname(ctx, params.Login).
// 			Return(nil, errors.New("not found")).
// 			Times(1)

// 		result, err := authService.Login(ctx, params)

// 		if err == nil {
// 			t.Fatal("Ожидалась ошибка, получен успех")
// 		}

// 		if result != nil {
// 			t.Fatal("Ожидался nil результат при ошибке")
// 		}

// 		expectedErrMsg := "неверный логин или пароль"
// 		if err.Error() != expectedErrMsg {
// 			t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedErrMsg, err.Error())
// 		}
// 	})
// }

// func TestAuthService_ValidateToken(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockUserRepo := mock.NewMockUser(ctrl)
// 	mockRoleRepo := mock.NewMockRole(ctrl)
// 	mockSessionRepo := mock.NewMockSession(ctrl)
// 	mockDeviceService := servicemock.NewMockDevice(ctrl)
// 	mockLoginHistoryRepo := mock.NewMockLoginHistory(ctrl)
// 	mockTokenManager := servicemock.NewMockTokenManager(ctrl)
// 	mockIDClient := NewMockIDClient()

// 	config := &config.Config{}

// 	authService := NewAuthService(
// 		config,
// 		mockUserRepo,
// 		mockRoleRepo,
// 		mockSessionRepo,
// 		mockDeviceService,
// 		mockLoginHistoryRepo,
// 		mockTokenManager,
// 		mockIDClient,
// 	)

// 	t.Run("валидный токен", func(t *testing.T) {
// 		token := "valid-token"
// 		userID := int64(1)
// 		roles := []string{"user", "admin"}

// 		mockTokenManager.EXPECT().
// 			ValidateToken(token).
// 			Return(&service.TokenPayload{
// 				UserID: userID,
// 				Roles:  roles,
// 				Exp:    time.Now().Add(time.Hour).Unix(),
// 			}, nil).
// 			Times(1)

// 		resultUserID, resultRoles, err := authService.ValidateToken(token)

// 		if err != nil {
// 			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
// 		}

// 		if resultUserID != userID {
// 			t.Errorf("Ожидался UserID %d, получен %d", userID, resultUserID)
// 		}

// 		if len(resultRoles) != len(roles) {
// 			t.Errorf("Ожидалось %d ролей, получено %d", len(roles), len(resultRoles))
// 		}
// 	})

// 	t.Run("невалидный токен", func(t *testing.T) {
// 		token := "invalid-token"

// 		mockTokenManager.EXPECT().
// 			ValidateToken(token).
// 			Return(nil, errors.New("invalid token")).
// 			Times(1)

// 		resultUserID, resultRoles, err := authService.ValidateToken(token)

// 		if err == nil {
// 			t.Fatal("Ожидалась ошибка, получен успех")
// 		}

// 		if resultUserID != 0 {
// 			t.Errorf("Ожидался UserID 0, получен %d", resultUserID)
// 		}

// 		if resultRoles != nil {
// 			t.Errorf("Ожидались nil роли, получены %v", resultRoles)
// 		}
// 	})
// }

// func TestAuthService_GetPublicKey(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	mockUserRepo := mock.NewMockUser(ctrl)
// 	mockRoleRepo := mock.NewMockRole(ctrl)
// 	mockSessionRepo := mock.NewMockSession(ctrl)
// 	mockDeviceService := servicemock.NewMockDevice(ctrl)
// 	mockLoginHistoryRepo := mock.NewMockLoginHistory(ctrl)
// 	mockTokenManager := servicemock.NewMockTokenManager(ctrl)
// 	mockIDClient := NewMockIDClient()

// 	config := &config.Config{}

// 	authService := NewAuthService(
// 		config,
// 		mockUserRepo,
// 		mockRoleRepo,
// 		mockSessionRepo,
// 		mockDeviceService,
// 		mockLoginHistoryRepo,
// 		mockTokenManager,
// 		mockIDClient,
// 	)

// 	t.Run("получение публичного ключа", func(t *testing.T) {
// 		expectedKey := []byte("test-public-key")

// 		mockTokenManager.EXPECT().
// 			GetPublicKey().
// 			Return(ed25519.PublicKey(expectedKey)).
// 			Times(1)

// 		result := authService.GetPublicKey()

// 		if result == nil {
// 			t.Fatal("Ожидался публичный ключ, получен nil")
// 		}

// 		if len(result) != len(expectedKey) {
// 			t.Errorf("Ожидалась длина ключа %d, получена %d", len(expectedKey), len(result))
// 		}
// 	})
// }

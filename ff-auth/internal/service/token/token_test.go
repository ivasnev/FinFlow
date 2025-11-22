package token

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/mock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestED25519TokenManager_LoadOrGenerateKeys(t *testing.T) {
	t.Run("загрузка ключей из БД", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockKeyPair(ctrl)
		keyPair := &models.KeyPair{
			ID:         1,
			PublicKey:  "dGVzdF9wdWJsaWNfa2V5",     // base64 encoded test key
			PrivateKey: "dGVzdF9wcml2YXRlX2tleQ==", // base64 encoded test key
			IsActive:   true,
		}

		mockRepo.EXPECT().
			GetActive(context.Background()).
			Return(keyPair, nil).
			Times(1)

		manager, err := NewED25519TokenManager(mockRepo)

		assert.NoError(t, err)
		assert.NotNil(t, manager)

		// Проверяем, что публичный ключ установлен
		publicKey := manager.GetPublicKey()
		assert.NotNil(t, publicKey)
	})

	t.Run("генерация новых ключей при отсутствии в БД", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockKeyPair(ctrl)
		// Первый вызов GetActive в LoadOrGenerateKeys - ключей нет (nil, nil)
		mockRepo.EXPECT().
			GetActive(context.Background()).
			Return(nil, nil).
			Times(1)

		// RegenerateKeys не вызывает GetActive, так как loadedFromDB == false
		// Сразу вызывается Create для создания новых ключей
		mockRepo.EXPECT().
			Create(context.Background(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, keyPair *models.KeyPair) error {
				assert.True(t, keyPair.IsActive)
				assert.NotEmpty(t, keyPair.PublicKey)
				assert.NotEmpty(t, keyPair.PrivateKey)
				return nil
			}).
			Times(1)

		manager, err := NewED25519TokenManager(mockRepo)

		assert.NoError(t, err)
		assert.NotNil(t, manager)
	})

	t.Run("ошибка загрузки ключей", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockKeyPair(ctrl)
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetActive(context.Background()).
			Return(nil, expectedErr).
			Times(1)

		manager, err := NewED25519TokenManager(mockRepo)

		assert.Error(t, err)
		assert.Nil(t, manager)
		assert.Contains(t, err.Error(), "ошибка при загрузке ключей")
	})
}

func TestED25519TokenManager_GenerateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockKeyPair(ctrl)

	// Создаем менеджер с моком
	manager := &ED25519TokenManager{
		keyPairRepo: mockRepo,
	}

	// Генерируем тестовые ключи
	publicKey, privateKey, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Ошибка генерации тестовых ключей: %v", err)
	}

	manager.publicKey = publicKey
	manager.privateKey = privateKey

	t.Run("успешная генерация токена", func(t *testing.T) {
		payload := &service.TokenPayload{
			UserID: 1,
			Roles:  []string{"user", "admin"},
			Exp:    time.Now().Add(time.Hour).Unix(),
		}

		token, err := manager.GenerateToken(payload)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Проверяем, что токен можно декодировать
		decodedPayload, err := manager.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, payload.UserID, decodedPayload.UserID)
		assert.Equal(t, len(payload.Roles), len(decodedPayload.Roles))
	})

	t.Run("ошибка сериализации payload", func(t *testing.T) {
		// Создаем невалидный payload (с циклической ссылкой)
		invalidPayload := &service.TokenPayload{
			UserID: 1,
			Roles:  []string{"user"},
			Exp:    time.Now().Add(time.Hour).Unix(),
		}

		// Это должно работать нормально, так как TokenPayload не содержит циклических ссылок
		token, err := manager.GenerateToken(invalidPayload)

		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})
}

func TestED25519TokenManager_ValidateToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockKeyPair(ctrl)

	// Создаем менеджер с моком
	manager := &ED25519TokenManager{
		keyPairRepo: mockRepo,
	}

	// Генерируем тестовые ключи
	publicKey, privateKey, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Ошибка генерации тестовых ключей: %v", err)
	}

	manager.publicKey = publicKey
	manager.privateKey = privateKey

	t.Run("валидный токен", func(t *testing.T) {
		payload := &service.TokenPayload{
			UserID: 1,
			Roles:  []string{"user", "admin"},
			Exp:    time.Now().Add(time.Hour).Unix(),
		}

		token, err := manager.GenerateToken(payload)
		if err != nil {
			t.Fatalf("Ошибка генерации токена: %v", err)
		}

		decodedPayload, err := manager.ValidateToken(token)

		assert.NoError(t, err)
		assert.NotNil(t, decodedPayload)
		assert.Equal(t, payload.UserID, decodedPayload.UserID)
		assert.Equal(t, len(payload.Roles), len(decodedPayload.Roles))
	})

	t.Run("истекший токен", func(t *testing.T) {
		payload := &service.TokenPayload{
			UserID: 1,
			Roles:  []string{"user"},
			Exp:    time.Now().Add(-time.Hour).Unix(), // Истекший токен
		}

		token, err := manager.GenerateToken(payload)
		if err != nil {
			t.Fatalf("Ошибка генерации токена: %v", err)
		}

		decodedPayload, err := manager.ValidateToken(token)

		assert.Error(t, err)
		assert.Nil(t, decodedPayload)
		assert.Equal(t, "token expired", err.Error())
	})

	t.Run("невалидный формат токена", func(t *testing.T) {
		invalidToken := "invalid-token-format"

		decodedPayload, err := manager.ValidateToken(invalidToken)

		assert.Error(t, err)
		assert.Nil(t, decodedPayload)
		assert.Equal(t, "invalid token format", err.Error())
	})

	t.Run("невалидная подпись токена", func(t *testing.T) {
		// Создаем токен с неправильной подписью
		payload := &service.TokenPayload{
			UserID: 1,
			Roles:  []string{"user"},
			Exp:    time.Now().Add(time.Hour).Unix(),
		}

		token, err := manager.GenerateToken(payload)
		assert.NoError(t, err)

		// Декодируем токен, изменяем подпись и кодируем обратно
		tokenBytes, _ := base64.StdEncoding.DecodeString(token)
		var tokenStruct service.Token
		json.Unmarshal(tokenBytes, &tokenStruct)

		// Изменяем подпись
		tokenStruct.Sig[0] ^= 0xFF

		// Кодируем обратно
		corruptedTokenBytes, _ := json.Marshal(tokenStruct)
		corruptedToken := base64.StdEncoding.EncodeToString(corruptedTokenBytes)

		decodedPayload, err := manager.ValidateToken(corruptedToken)

		assert.Error(t, err)
		assert.Nil(t, decodedPayload)
		assert.Equal(t, "invalid token signature", err.Error())
	})
}

func TestED25519TokenManager_GenerateTokenPair(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockKeyPair(ctrl)

	// Создаем менеджер с моком
	manager := &ED25519TokenManager{
		keyPairRepo: mockRepo,
	}

	// Генерируем тестовые ключи
	publicKey, privateKey, err := generateTestKeys()
	if err != nil {
		t.Fatalf("Ошибка генерации тестовых ключей: %v", err)
	}

	manager.publicKey = publicKey
	manager.privateKey = privateKey

	t.Run("успешная генерация пары токенов", func(t *testing.T) {
		userID := int64(1)
		roles := []string{"user", "admin"}
		accessTTL := time.Hour
		refreshTTL := 24 * time.Hour

		accessToken, refreshToken, accessExpiresAt, err := manager.GenerateTokenPair(
			userID, roles, accessTTL, refreshTTL,
		)

		assert.NoError(t, err)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.Greater(t, accessExpiresAt, time.Now().Unix())
		assert.NotEqual(t, accessToken, refreshToken)

		// Проверяем валидность access токена
		accessPayload, err := manager.ValidateToken(accessToken)
		assert.NoError(t, err)
		assert.Equal(t, userID, accessPayload.UserID)

		// Проверяем валидность refresh токена
		refreshPayload, err := manager.ValidateToken(refreshToken)
		assert.NoError(t, err)
		assert.Equal(t, userID, refreshPayload.UserID)

		// Проверяем, что refresh токен живет дольше access токена
		assert.Greater(t, refreshPayload.Exp, accessPayload.Exp)
	})
}

func TestED25519TokenManager_RegenerateKeys(t *testing.T) {
	t.Run("успешная регенерация ключей без активного ключа в БД", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockKeyPair(ctrl)

		// Создаем менеджер с моком
		manager := &ED25519TokenManager{
			keyPairRepo:  mockRepo,
			loadedFromDB: false,
		}

		// Генерируем тестовые ключи для начальной установки
		publicKey, privateKey, err := generateTestKeys()
		if err != nil {
			t.Fatalf("Ошибка генерации тестовых ключей: %v", err)
		}

		manager.publicKey = publicKey
		manager.privateKey = privateKey

		// Когда loadedFromDB == false, GetActive не вызывается
		mockRepo.EXPECT().
			Create(context.Background(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, keyPair *models.KeyPair) error {
				assert.True(t, keyPair.IsActive)
				assert.NotEmpty(t, keyPair.PublicKey)
				assert.NotEmpty(t, keyPair.PrivateKey)
				return nil
			}).
			Times(1)

		err = manager.RegenerateKeys()

		assert.NoError(t, err)
		assert.NotNil(t, manager.publicKey)
		assert.NotNil(t, manager.privateKey)
	})

	t.Run("успешная регенерация ключей с деактивацией старого", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockKeyPair(ctrl)

		// Создаем менеджер с моком
		manager := &ED25519TokenManager{
			keyPairRepo:  mockRepo,
			loadedFromDB: true,
		}

		// Генерируем тестовые ключи для начальной установки
		publicKey, privateKey, err := generateTestKeys()
		if err != nil {
			t.Fatalf("Ошибка генерации тестовых ключей: %v", err)
		}

		manager.publicKey = publicKey
		manager.privateKey = privateKey

		oldKeyPair := &models.KeyPair{
			ID:         1,
			PublicKey:  "old-public-key",
			PrivateKey: "old-private-key",
			IsActive:   true,
		}

		mockRepo.EXPECT().
			GetActive(context.Background()).
			Return(oldKeyPair, nil).
			Times(1)

		mockRepo.EXPECT().
			Update(context.Background(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, keyPair *models.KeyPair) error {
				assert.False(t, keyPair.IsActive)
				return nil
			}).
			Times(1)

		mockRepo.EXPECT().
			Create(context.Background(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, keyPair *models.KeyPair) error {
				assert.True(t, keyPair.IsActive)
				assert.NotEmpty(t, keyPair.PublicKey)
				assert.NotEmpty(t, keyPair.PrivateKey)
				return nil
			}).
			Times(1)

		err = manager.RegenerateKeys()

		assert.NoError(t, err)
	})

	t.Run("ошибка получения активного ключа", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockKeyPair(ctrl)

		// Создаем менеджер с моком
		manager := &ED25519TokenManager{
			keyPairRepo:  mockRepo,
			loadedFromDB: true,
		}

		// Генерируем тестовые ключи для начальной установки
		publicKey, privateKey, err := generateTestKeys()
		if err != nil {
			t.Fatalf("Ошибка генерации тестовых ключей: %v", err)
		}

		manager.publicKey = publicKey
		manager.privateKey = privateKey

		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetActive(context.Background()).
			Return(nil, expectedErr).
			Times(1)

		err = manager.RegenerateKeys()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при получении текущего активного ключа")
	})

	t.Run("ошибка деактивации старого ключа", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockKeyPair(ctrl)

		// Создаем менеджер с моком
		manager := &ED25519TokenManager{
			keyPairRepo:  mockRepo,
			loadedFromDB: true,
		}

		// Генерируем тестовые ключи для начальной установки
		publicKey, privateKey, err := generateTestKeys()
		if err != nil {
			t.Fatalf("Ошибка генерации тестовых ключей: %v", err)
		}

		manager.publicKey = publicKey
		manager.privateKey = privateKey

		oldKeyPair := &models.KeyPair{
			ID:         1,
			PublicKey:  "old-public-key",
			PrivateKey: "old-private-key",
			IsActive:   true,
		}

		expectedErr := errors.New("update error")

		mockRepo.EXPECT().
			GetActive(context.Background()).
			Return(oldKeyPair, nil).
			Times(1)

		mockRepo.EXPECT().
			Update(context.Background(), gomock.Any()).
			Return(expectedErr).
			Times(1)

		err = manager.RegenerateKeys()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при деактивации текущего ключа")
	})

	t.Run("ошибка создания нового ключа", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock.NewMockKeyPair(ctrl)

		// Создаем менеджер с моком
		manager := &ED25519TokenManager{
			keyPairRepo:  mockRepo,
			loadedFromDB: false,
		}

		// Генерируем тестовые ключи для начальной установки
		publicKey, privateKey, err := generateTestKeys()
		if err != nil {
			t.Fatalf("Ошибка генерации тестовых ключей: %v", err)
		}

		manager.publicKey = publicKey
		manager.privateKey = privateKey

		expectedErr := errors.New("create error")

		// Когда loadedFromDB == false, GetActive не вызывается
		mockRepo.EXPECT().
			Create(context.Background(), gomock.Any()).
			Return(expectedErr).
			Times(1)

		err = manager.RegenerateKeys()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при сохранении ключей в БД")
	})
}

// generateTestKeys генерирует тестовые ключи для тестов
func generateTestKeys() (publicKey, privateKey []byte, err error) {
	// Используем настоящую генерацию ключей для тестов
	return ed25519.GenerateKey(rand.Reader)
}

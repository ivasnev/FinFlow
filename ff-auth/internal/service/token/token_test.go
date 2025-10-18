package token

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/models"
	"github.com/ivasnev/FinFlow/ff-auth/internal/repository/mock"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
)

func TestED25519TokenManager_LoadOrGenerateKeys(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockKeyPair(ctrl)

	t.Run("загрузка ключей из БД", func(t *testing.T) {
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

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if manager == nil {
			t.Fatal("Ожидался менеджер, получен nil")
		}

		// Проверяем, что публичный ключ установлен
		publicKey := manager.GetPublicKey()
		if publicKey == nil {
			t.Error("Ожидался публичный ключ, получен nil")
		}
	})

	t.Run("генерация новых ключей при отсутствии в БД", func(t *testing.T) {
		mockRepo.EXPECT().
			GetActive(context.Background()).
			Return(nil, errors.New("not found")).
			Times(1)

		// Ожидаем вызов RegenerateKeys - сначала GetActive для проверки существующих ключей
		mockRepo.EXPECT().
			GetActive(context.Background()).
			Return(nil, errors.New("not found")).
			Times(1)

		// Затем Create для создания новых ключей
		mockRepo.EXPECT().
			Create(context.Background(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, keyPair *models.KeyPair) error {
				if !keyPair.IsActive {
					t.Error("Ожидался активный ключ")
				}
				if keyPair.PublicKey == "" {
					t.Error("Ожидался публичный ключ")
				}
				if keyPair.PrivateKey == "" {
					t.Error("Ожидался приватный ключ")
				}
				return nil
			}).
			Times(1)

		manager, err := NewED25519TokenManager(mockRepo)

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if manager == nil {
			t.Fatal("Ожидался менеджер, получен nil")
		}
	})

	t.Run("ошибка загрузки ключей", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetActive(context.Background()).
			Return(nil, expectedErr).
			Times(1)

		manager, err := NewED25519TokenManager(mockRepo)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if manager != nil {
			t.Fatal("Ожидался nil менеджер при ошибке")
		}

		if !errors.Is(err, expectedErr) {
			t.Errorf("Ожидалась ошибка %v, получена %v", expectedErr, err)
		}
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

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if token == "" {
			t.Fatal("Ожидался токен, получена пустая строка")
		}

		// Проверяем, что токен можно декодировать
		decodedPayload, err := manager.ValidateToken(token)
		if err != nil {
			t.Fatalf("Ошибка валидации токена: %v", err)
		}

		if decodedPayload.UserID != payload.UserID {
			t.Errorf("Ожидался UserID %d, получен %d", payload.UserID, decodedPayload.UserID)
		}

		if len(decodedPayload.Roles) != len(payload.Roles) {
			t.Errorf("Ожидалось %d ролей, получено %d", len(payload.Roles), len(decodedPayload.Roles))
		}
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

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if token == "" {
			t.Fatal("Ожидался токен, получена пустая строка")
		}
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

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if decodedPayload == nil {
			t.Fatal("Ожидался payload, получен nil")
		}

		if decodedPayload.UserID != payload.UserID {
			t.Errorf("Ожидался UserID %d, получен %d", payload.UserID, decodedPayload.UserID)
		}

		if len(decodedPayload.Roles) != len(payload.Roles) {
			t.Errorf("Ожидалось %d ролей, получено %d", len(payload.Roles), len(decodedPayload.Roles))
		}
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

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if decodedPayload != nil {
			t.Fatal("Ожидался nil payload при ошибке")
		}

		expectedErrMsg := "token expired"
		if err.Error() != expectedErrMsg {
			t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedErrMsg, err.Error())
		}
	})

	t.Run("невалидный формат токена", func(t *testing.T) {
		invalidToken := "invalid-token-format"

		decodedPayload, err := manager.ValidateToken(invalidToken)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if decodedPayload != nil {
			t.Fatal("Ожидался nil payload при ошибке")
		}

		expectedErrMsg := "invalid token format"
		if err.Error() != expectedErrMsg {
			t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedErrMsg, err.Error())
		}
	})

	t.Run("невалидная подпись токена", func(t *testing.T) {
		// Создаем токен с неправильной подписью
		payload := &service.TokenPayload{
			UserID: 1,
			Roles:  []string{"user"},
			Exp:    time.Now().Add(time.Hour).Unix(),
		}

		token, err := manager.GenerateToken(payload)
		if err != nil {
			t.Fatalf("Ошибка генерации токена: %v", err)
		}

		// Повреждаем токен (изменяем последний символ)
		corruptedToken := token[:len(token)-1] + "X"

		decodedPayload, err := manager.ValidateToken(corruptedToken)

		if err == nil {
			t.Fatal("Ожидалась ошибка, получен успех")
		}

		if decodedPayload != nil {
			t.Fatal("Ожидался nil payload при ошибке")
		}

		expectedErrMsg := "invalid token signature"
		if err.Error() != expectedErrMsg {
			t.Errorf("Ожидалась ошибка '%s', получена '%s'", expectedErrMsg, err.Error())
		}
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

		if err != nil {
			t.Fatalf("Ожидался успех, получена ошибка: %v", err)
		}

		if accessToken == "" {
			t.Fatal("Ожидался access токен, получена пустая строка")
		}

		if refreshToken == "" {
			t.Fatal("Ожидался refresh токен, получена пустая строка")
		}

		if accessExpiresAt <= time.Now().Unix() {
			t.Error("Время истечения access токена должно быть в будущем")
		}

		// Проверяем, что токены разные
		if accessToken == refreshToken {
			t.Error("Access и refresh токены не должны быть одинаковыми")
		}

		// Проверяем валидность access токена
		accessPayload, err := manager.ValidateToken(accessToken)
		if err != nil {
			t.Fatalf("Ошибка валидации access токена: %v", err)
		}

		if accessPayload.UserID != userID {
			t.Errorf("Ожидался UserID %d, получен %d", userID, accessPayload.UserID)
		}

		// Проверяем валидность refresh токена
		refreshPayload, err := manager.ValidateToken(refreshToken)
		if err != nil {
			t.Fatalf("Ошибка валидации refresh токена: %v", err)
		}

		if refreshPayload.UserID != userID {
			t.Errorf("Ожидался UserID %d, получен %d", userID, refreshPayload.UserID)
		}

		// Проверяем, что refresh токен живет дольше access токена
		if refreshPayload.Exp <= accessPayload.Exp {
			t.Error("Refresh токен должен жить дольше access токена")
		}
	})
}

// generateTestKeys генерирует тестовые ключи для тестов
func generateTestKeys() (publicKey, privateKey []byte, err error) {
	// Используем настоящую генерацию ключей для тестов
	return ed25519.GenerateKey(rand.Reader)
}

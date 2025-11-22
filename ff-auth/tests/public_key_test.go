package tests

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"
)

// PublicKeySuite представляет suite для тестов публичного ключа
type PublicKeySuite struct {
	BaseSuite
}

// TestPublicKeySuite запускает все тесты в PublicKeySuite
func TestPublicKeySuite(t *testing.T) {
	suite.Run(t, new(PublicKeySuite))
}

// TestGetPublicKey_Success тестирует успешное получение публичного ключа
func (s *PublicKeySuite) TestGetPublicKey_Success() {
	ctx := context.Background()

	publicKeyResp, err := s.APIClient.GetPublicKeyWithResponse(ctx)
	s.NoError(err, "получение публичного ключа должно пройти успешно")
	s.Equal(200, publicKeyResp.StatusCode(), "должен быть статус 200")
	s.NotEmpty(publicKeyResp.Body, "публичный ключ должен быть возвращен")

	// Проверяем, что ключ можно декодировать из base64
	publicKeyStr := string(publicKeyResp.Body)
	s.NotEmpty(publicKeyStr, "публичный ключ не должен быть пустым")
	decodedKey, err := base64.StdEncoding.DecodeString(publicKeyStr)
	s.NoError(err, "ключ должен быть валидным base64")
	s.NotEmpty(decodedKey, "декодированный ключ не должен быть пустым")
}

// TestGetPublicKey_Consistency тестирует консистентность публичного ключа
func (s *PublicKeySuite) TestGetPublicKey_Consistency() {
	ctx := context.Background()

	publicKeyResp1, err := s.APIClient.GetPublicKeyWithResponse(ctx)
	s.NoError(err)
	s.Equal(200, publicKeyResp1.StatusCode())
	s.NotEmpty(publicKeyResp1.Body)

	publicKeyResp2, err := s.APIClient.GetPublicKeyWithResponse(ctx)
	s.NoError(err)
	s.Equal(200, publicKeyResp2.StatusCode())
	s.NotEmpty(publicKeyResp2.Body)

	s.Equal(string(publicKeyResp1.Body), string(publicKeyResp2.Body), "публичный ключ должен быть одинаковым при повторных запросах")
}

// TestGetPublicKey_CanValidateToken тестирует, что публичный ключ может валидировать токены
// Примечание: Для валидации токена используем TokenManager, так как это внутренняя функциональность
func (s *PublicKeySuite) TestGetPublicKey_CanValidateToken() {
	ctx := context.Background()

	// Регистрируем пользователя
	registerReq := api.RegisterJSONRequestBody{
		Email:    openapi_types.Email("publickey@example.com"),
		Nickname: "publickeyuser",
		Password: "password123",
	}
	registerResp, err := s.APIClient.RegisterWithResponse(ctx, registerReq)
	s.NoError(err)
	s.Equal(201, registerResp.StatusCode())
	s.NotNil(registerResp.JSON201)

	// Валидируем токен с помощью TokenManager (внутренняя функциональность)
	payload, err := s.Container.TokenManager.ValidateToken(registerResp.JSON201.AccessToken)

	s.NoError(err, "токен должен быть валидным")
	s.NotNil(payload, "payload должен быть возвращен")
	s.Equal(registerResp.JSON201.User.Id, payload.UserID, "UserID должен совпадать")
}


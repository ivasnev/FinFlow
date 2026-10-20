package tests

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"gorm.io/gorm"
)

// TestRefresh_AfterAccessExpiry проверяет срок сессии после регистрации, входа и ротации токенов.
func (s *RefreshSuite) TestRefresh_AfterAccessExpiry() {
	ctx := context.Background()
	previousTTL := s.Config.Auth.AccessTokenDuration
	s.Config.Auth.AccessTokenDuration = -1
	defer func() { s.Config.Auth.AccessTokenDuration = previousTTL }()
	s.MockServer.Expect(http.MethodPost, "/api/v1/internal/users/register").Return("ff_id_service/register_user_response_success.json").HTTPCode(http.StatusCreated)
	registered, err := s.APIClient.RegisterWithResponse(ctx, api.RegisterJSONRequestBody{Email: openapi_types.Email("expiry@example.com"), Nickname: "expiryuser", Password: "password123"})
	s.Require().NoError(err)
	s.Require().Equal(201, registered.StatusCode())
	s.Require().NotNil(registered.JSON201)
	assertExpiry := func(refresh string, accessExpiry time.Time) int64 {
		s.T().Helper()
		s.Require().True(accessExpiry.Before(time.Now()), "access must already be expired in this scenario")
		data, err := base64.StdEncoding.DecodeString(refresh)
		s.Require().NoError(err)
		var envelope auth.Token
		s.Require().NoError(json.Unmarshal(data, &envelope))
		s.Require().True(ed25519.Verify(s.Container.TokenManager.GetPublicKey(), envelope.Payload, envelope.Sig))
		var payload struct {
			Exp int64 `json:"exp"`
		}
		s.Require().NoError(json.Unmarshal(envelope.Payload, &payload))
		var session struct{ ExpiresAt time.Time }
		s.Require().NoError(s.DBContainer.DB.Table("sessions").Select("expires_at").Where("refresh_token = ?", refresh).Take(&session).Error)
		s.Equal(payload.Exp, session.ExpiresAt.Unix())
		s.Greater(session.ExpiresAt.Unix(), time.Now().Add(6*24*time.Hour).Unix())
		return payload.Exp
	}
	assertExpiry(registered.JSON201.RefreshToken, registered.JSON201.ExpiresAt)
	refreshed, err := s.APIClient.RefreshTokenWithResponse(ctx, api.RefreshTokenJSONRequestBody{RefreshToken: registered.JSON201.RefreshToken})
	s.Require().NoError(err)
	s.Require().Equal(200, refreshed.StatusCode())
	s.Require().NotNil(refreshed.JSON200)
	assertExpiry(refreshed.JSON200.RefreshToken, refreshed.JSON200.ExpiresAt)
	rotated, err := s.APIClient.RefreshTokenWithResponse(ctx, api.RefreshTokenJSONRequestBody{RefreshToken: refreshed.JSON200.RefreshToken})
	s.Require().NoError(err)
	s.Require().Equal(200, rotated.StatusCode())
	s.Require().NotNil(rotated.JSON200)
	assertExpiry(rotated.JSON200.RefreshToken, rotated.JSON200.ExpiresAt)
	login, err := s.APIClient.LoginWithResponse(ctx, api.LoginJSONRequestBody{Login: "expiry@example.com", Password: "password123"})
	s.Require().NoError(err)
	s.Require().Equal(200, login.StatusCode())
	s.Require().NotNil(login.JSON200)
	expectedExpiry := assertExpiry(login.JSON200.RefreshToken, login.JSON200.ExpiresAt)
	// Проверяем реальные файлы миграции на существующих строках и при другом часовом поясе БД.
	up, err := os.ReadFile("../internal/repository/migrations/000003_session_expiry_timezone.up.sql")
	s.Require().NoError(err)
	down, err := os.ReadFile("../internal/repository/migrations/000003_session_expiry_timezone.down.sql")
	s.Require().NoError(err)
	s.Require().NoError(s.DBContainer.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET LOCAL TIME ZONE 'Pacific/Auckland'").Error; err != nil {
			return err
		}
		for _, migration := range [][]byte{down, up} {
			if err := tx.Exec(string(migration)).Error; err != nil {
				return err
			}
		}
		var session struct{ ExpiresAt time.Time }
		if err := tx.Table("sessions").Select("expires_at").Where("refresh_token = ?", login.JSON200.RefreshToken).Take(&session).Error; err != nil {
			return err
		}
		if session.ExpiresAt.Unix() != expectedExpiry {
			return fmt.Errorf("migration changed expiry: %d want %d", session.ExpiresAt.Unix(), expectedExpiry)
		}
		return nil
	}))
}

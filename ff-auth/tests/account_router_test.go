package tests

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/account"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	"github.com/ivasnev/FinFlow/ff-auth/tests/mockserver"
	"github.com/stretchr/testify/require"
)

// TestAccountProductionRouter exercises the real generated router and middleware, including typed scope constants.
func TestAccountProductionRouter(t *testing.T) {
	database := setupTestDB(t)
	t.Cleanup(func() { teardownTestDB(t, database) })
	migration, err := os.ReadFile("../internal/repository/migrations/000005_account_security.up.sql")
	require.NoError(t, err)
	require.NoError(t, database.DB.Exec(string(migration)).Error)
	idServer := mockserver.NewMockServer()
	t.Cleanup(idServer.Stop)
	idServer.Expect(http.MethodPost, "/api/v1/internal/users/register").Return("ff_id_service/register_user_response_success.json").HTTPCode(201)
	idClient, err := ffid.NewAdapter(idServer.GetBaseURL(), http.DefaultClient)
	require.NoError(t, err)
	cfg := &config.Config{}
	cfg.Auth.PasswordHashCost = 4
	cfg.Auth.AccessTokenDuration = 15
	cfg.Auth.RefreshTokenDuration = 10080
	c, err := createTestContainer(t, cfg, gin.New(), database.DB, idClient)
	require.NoError(t, err)
	session, err := c.AuthService.Register(context.Background(), service.RegisterParams{Email: "test@example.com", Nickname: "testuser", Password: "password123"})
	require.NoError(t, err)
	manager, err := account.New(database.DB, account.Config{Local: true, Secret: strings.Repeat("x", 40), AppURL: "http://127.0.0.1:3000", SMTPAddr: "127.0.0.1:1", From: "noreply@finflow.test", CaptchaProvider: "image", PasswordMin: 8, HashCost: 4, AccessTTL: time.Minute, RefreshTTL: time.Hour, MailDaily: 100, MailMonthly: 2000}, c.TokenManager, idClient)
	require.NoError(t, err)
	c.Account = manager
	c.ServerHandler.Account = manager
	c.RegisterRoutes()
	server := httptest.NewServer(c.Router)
	t.Cleanup(server.Close)
	probe := func(method, path, body, token string) int {
		t.Helper()
		req, err := http.NewRequest(method, server.URL+"/api/v1"+path, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		return resp.StatusCode
	}
	require.Equal(t, 401, probe("GET", "/sessions", "", session.AccessToken))
	require.Equal(t, 401, probe("GET", "/login-history", "", session.AccessToken))
	require.Equal(t, 401, probe("POST", "/auth/introspect", "", session.AccessToken))
	require.Equal(t, 401, probe("GET", "/sessions", "", session.RefreshToken))
	// Seed the exact verified-email/live-session relationship to test the positive and revoked cases independently of the signup UI.
	require.NoError(t, database.DB.Exec(`INSERT INTO account_security(user_id,verified_email) VALUES(?,?)`, session.User.Id, session.User.Email).Error)
	hash := sha256.Sum256([]byte(session.AccessToken))
	require.NoError(t, database.DB.Exec(`INSERT INTO account_access(token_hash,session_id,expires_at) SELECT ?,id,? FROM sessions WHERE refresh_token=?`, hex.EncodeToString(hash[:]), session.ExpiresAt, session.RefreshToken).Error)
	require.Equal(t, 200, probe("GET", "/sessions", "", session.AccessToken))
	require.Equal(t, 200, probe("POST", "/auth/introspect", "", session.AccessToken))
	require.Equal(t, 400, probe("PATCH", "/users/me", `{"email":"bypass@example.test"}`, session.AccessToken))
	require.Equal(t, 400, probe("PATCH", "/users/me", `{"password":"bypass-password"}`, session.AccessToken))
	require.NoError(t, database.DB.Exec(`DELETE FROM sessions WHERE user_id=?`, session.User.Id).Error)
	require.Equal(t, 401, probe("GET", "/sessions", "", session.AccessToken))
	require.Equal(t, 401, probe("POST", "/auth/introspect", "", session.AccessToken))
}

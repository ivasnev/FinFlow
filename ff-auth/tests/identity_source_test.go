package tests

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// identityTransport добавляет подписанный тестовый билет к реальным межсервисным HTTP-запросам.
type identityTransport struct{ ticket string }

// RoundTrip передаёт запрос в настоящий ff-id через штатную проверку TVM.
func (t identityTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("X-FF-Service-Ticket", t.ticket)
	return http.DefaultTransport.RoundTrip(clone)
}

// TestIdentitySource проверяет два настоящих HTTP-сервиса с отдельными PostgreSQL-базами.
func TestIdentitySource(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	database := setupTestDB(t)
	t.Cleanup(func() { teardownTestDB(t, database) })
	require.NoError(t, database.DB.Exec("CREATE DATABASE ff_identity_source").Error)
	host, err := database.Container.Host(ctx)
	require.NoError(t, err)
	port, err := database.Container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err)
	identityDB, err := gorm.Open(postgres.Open(fmt.Sprintf("host=%s port=%s user=postgres password=postgres dbname=ff_identity_source sslmode=disable", host, port.Port())), &gorm.Config{})
	require.NoError(t, err)
	connection, err := identityDB.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = connection.Close() })
	migrations, err := filepath.Glob("../../ff-id/internal/repository/migrations/*.up.sql")
	require.NoError(t, err)
	require.NotEmpty(t, migrations)
	for _, path := range migrations {
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		require.NoError(t, identityDB.Exec(string(data)).Error)
	}
	// Проверяем реальную миграцию на прежних ограничениях auth.
	require.NoError(t, database.DB.Exec("ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE(email); ALTER TABLE users ADD CONSTRAINT users_nickname_key UNIQUE(nickname)").Error)
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	keys := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"public_key": base64.StdEncoding.EncodeToString(pub)})
	}))
	t.Cleanup(keys.Close)
	payload := struct {
		From     int    `json:"from"`
		To       int    `json:"to"`
		TTL      int64  `json:"ttl"`
		Metadata string `json:"metadata"`
	}{4, 2, time.Now().Add(time.Hour).Unix(), "{}"}
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	envelope, err := json.Marshal(struct {
		Payload   interface{} `json:"payload"`
		Signature string      `json:"signature"`
	}{payload, base64.StdEncoding.EncodeToString(ed25519.Sign(priv, data))})
	require.NoError(t, err)
	ticket := "serv:NA==:" + base64.StdEncoding.EncodeToString(envelope)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	idPort := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	idURL := fmt.Sprintf("http://127.0.0.1:%d", idPort)
	idClient, err := ffid.NewAdapter(idURL, &http.Client{Transport: identityTransport{ticket}, Timeout: 5 * time.Second})
	require.NoError(t, err)
	cfg := &config.Config{}
	cfg.Auth.AccessTokenDuration = 15
	cfg.Auth.RefreshTokenDuration = 10080
	cfg.Auth.PasswordHashCost = 4
	c, err := createTestContainer(t, cfg, gin.New(), database.DB, idClient)
	require.NoError(t, err)
	c.RegisterRoutes()
	// This fixture isolates the legacy identity-source contract. Full account lifecycle and live-grant checks have separate integration tests.
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/introspect" {
			payload, err := c.TokenManager.ValidateToken(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
			if err != nil {
				w.WriteHeader(401)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"active": true, "user_id": payload.UserID})
			return
		}
		c.Router.ServeHTTP(w, r)
	}))
	t.Cleanup(authServer.Close)
	authURL, err := url.Parse(authServer.URL)
	require.NoError(t, err)
	binary := filepath.Join(t.TempDir(), "ff-id")
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/app")
	build.Dir = "../../ff-id"
	output, err := build.CombinedOutput()
	require.NoError(t, err, "ff-id build failed: %s", output)
	log, err := os.CreateTemp(t.TempDir(), "ff-id-*.log")
	require.NoError(t, err)
	defer log.Close()
	command := exec.CommandContext(ctx, binary)
	command.Dir = "../../ff-id"
	command.Stdout = log
	command.Stderr = log
	command.Env = append(os.Environ(), "POSTGRES_HOST="+host, "POSTGRES_PORT="+port.Port(), "POSTGRES_DB=ff_identity_source", "POSTGRES_USER=postgres", "POSTGRES_PASSWORD=postgres", "SERVER_PORT="+strconv.Itoa(idPort), "TVM_SERVICE_ID=2", "AUTH_SERVICE_ID=4", "TVM_BASE_URL="+keys.URL, "AUTH_CLIENT_HOST=http://"+authURL.Hostname(), "AUTH_CLIENT_PORT="+authURL.Port(), "GIN_MODE=release")
	require.NoError(t, command.Start())
	wait := make(chan error, 1)
	go func() { wait <- command.Wait() }()
	stop := sync.OnceFunc(func() {
		_ = command.Process.Signal(os.Interrupt)
		select {
		case <-wait:
		case <-time.After(6 * time.Second):
			_ = command.Process.Kill()
			<-wait
		}
	})
	t.Cleanup(stop)
	httpClient := &http.Client{Timeout: 5 * time.Second}
	require.Eventually(t, func() bool {
		resp, err := httpClient.Get(idURL + "/api/v1/internal/identity?login=ready")
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == 401
	}, 10*time.Second, 50*time.Millisecond)
	client, err := api.NewClientWithResponses(authServer.URL + "/api/v1")
	require.NoError(t, err)
	registered, err := client.RegisterWithResponse(ctx, api.RegisterJSONRequestBody{Email: "old@example.com", Nickname: "old", Password: "password123"})
	require.NoError(t, err)
	require.Equal(t, 201, registered.StatusCode())
	require.NotNil(t, registered.JSON201)
	access := registered.JSON201.AccessToken
	request := func(method, path, body string, token bool) int {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, method, idURL+path, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		if token {
			req.Header.Set("Authorization", "Bearer "+access)
		}
		resp, err := httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		return resp.StatusCode
	}
	require.Equal(t, 401, request("PUT", "/api/v1/internal/identity/1", `{"nickname":"attacker"}`, false))
	require.Equal(t, 400, request("PATCH", "/api/v1/users/me", `{"email":"id@example.com","nickname":"id-new"}`, true))
	initialEmail, initialNickname := "id@example.com", "id-new"
	_, err = idClient.UpdateIdentity(ctx, 1, &initialEmail, &initialNickname)
	require.NoError(t, err)
	login := func(name string, want int) {
		t.Helper()
		resp, err := client.LoginWithResponse(ctx, api.LoginJSONRequestBody{Login: name, Password: "password123"})
		require.NoError(t, err)
		require.Equal(t, want, resp.StatusCode())
	}
	login("old", 401)
	login("old@example.com", 401)
	login("id-new", 200)
	login("id@example.com", 200)
	refreshed, err := client.RefreshTokenWithResponse(ctx, api.RefreshTokenJSONRequestBody{RefreshToken: registered.JSON201.RefreshToken})
	require.NoError(t, err)
	require.Equal(t, 200, refreshed.StatusCode())
	require.NotNil(t, refreshed.JSON200)
	require.Equal(t, "id-new", refreshed.JSON200.User.Nickname)

	newEmail, newNickname := openapi_types.Email("auth@example.com"), "auth-new"
	edit := func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+access)
		return nil
	}
	update, err := client.UpdateUserWithResponse(ctx, api.UpdateUserJSONRequestBody{Email: &newEmail, Nickname: &newNickname}, edit)
	require.NoError(t, err)
	require.Equal(t, 200, update.StatusCode())
	profile, err := idClient.GetIdentity(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, newNickname, profile.Nickname)
	require.Equal(t, string(newEmail), profile.Email)
	login("id-new", 401)
	login("id@example.com", 401)
	login("auth-new", 200)
	login("auth@example.com", 200)
	up, err := os.ReadFile("../internal/repository/migrations/000004_identity_source.up.sql")
	require.NoError(t, err)
	require.NoError(t, database.DB.Exec(string(up)).Error)
	reused, err := client.RegisterWithResponse(ctx, api.RegisterJSONRequestBody{Email: "old@example.com", Nickname: "old", Password: "password123"})
	require.NoError(t, err)
	require.Equal(t, 201, reused.StatusCode())
	// Сценарий: ошибка ff-id не меняет пароль в auth.
	require.NoError(t, identityDB.Exec(`CREATE FUNCTION reject_identity() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected identity failure'; END; $$; CREATE TRIGGER reject_identity BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION reject_identity();`).Error)
	newPassword := "newpassword123"
	partialEmail := openapi_types.Email("partial@example.com")
	refused, err := client.UpdateUserWithResponse(ctx, api.UpdateUserJSONRequestBody{Email: &partialEmail, Password: &newPassword}, edit)
	require.NoError(t, err)
	require.Equal(t, 503, refused.StatusCode())
	login("auth-new", 200)
	require.NoError(t, identityDB.Exec("DROP TRIGGER reject_identity ON users; DROP FUNCTION reject_identity()").Error)
	// Сценарий: профиль уже изменён в ff-id, но локальная смена пароля не прошла; идентификаторы остаются согласованными.
	require.NoError(t, database.DB.Exec(`CREATE FUNCTION reject_password() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected password failure'; END; $$; CREATE TRIGGER reject_password BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION reject_password();`).Error)
	partial, err := client.UpdateUserWithResponse(ctx, api.UpdateUserJSONRequestBody{Email: &partialEmail, Password: &newPassword}, edit)
	require.NoError(t, err)
	require.Equal(t, 500, partial.StatusCode())
	login("partial@example.com", 200)
	login("auth@example.com", 401)
	require.NoError(t, database.DB.Exec("DROP TRIGGER reject_password ON users; DROP FUNCTION reject_password()").Error)
	retried, err := client.UpdateUserWithResponse(ctx, api.UpdateUserJSONRequestBody{Password: &newPassword}, edit)
	require.NoError(t, err)
	require.Equal(t, 200, retried.StatusCode())
	finalLogin, err := client.LoginWithResponse(ctx, api.LoginJSONRequestBody{Login: "partial@example.com", Password: newPassword})
	require.NoError(t, err)
	require.Equal(t, 200, finalLogin.StatusCode())
	// Сценарий: nickname с символом @ не смешивается с email другого пользователя.
	emailOwner, err := client.RegisterWithResponse(ctx, api.RegisterJSONRequestBody{Email: "taken@example.com", Nickname: "email-owner", Password: "password123"})
	require.NoError(t, err)
	require.Equal(t, 201, emailOwner.StatusCode())
	require.NotNil(t, emailOwner.JSON201)
	nicknameOwner, err := client.RegisterWithResponse(ctx, api.RegisterJSONRequestBody{Email: "nickname-owner@example.com", Nickname: "taken@example.com", Password: "password123"})
	require.NoError(t, err)
	require.Equal(t, 201, nicknameOwner.StatusCode())
	require.NotNil(t, nicknameOwner.JSON201)
	byNickname, err := client.GetUserByNicknameWithResponse(ctx, "taken@example.com")
	require.NoError(t, err)
	require.Equal(t, 200, byNickname.StatusCode())
	require.NotNil(t, byNickname.JSON200)
	require.Equal(t, nicknameOwner.JSON201.User.Id, byNickname.JSON200.Id)
	byLogin, err := client.LoginWithResponse(ctx, api.LoginJSONRequestBody{Login: "taken@example.com", Password: "password123"})
	require.NoError(t, err)
	require.Equal(t, 200, byLogin.StatusCode())
	require.NotNil(t, byLogin.JSON200)
	require.Equal(t, emailOwner.JSON201.User.Id, byLogin.JSON200.User.Id)
	// Сценарий: гонка регистрации одной пары email/nickname оставляет ровно одну учётную запись.
	results := make(chan struct {
		status int
		err    error
	}, 8)
	var registrations sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 8; i++ {
		registrations.Add(1)
		go func() {
			defer registrations.Done()
			<-start
			response, err := client.RegisterWithResponse(ctx, api.RegisterJSONRequestBody{Email: "race@example.com", Nickname: "race", Password: "password123"})
			status := 0
			if err == nil {
				status = response.StatusCode()
			}
			results <- struct {
				status int
				err    error
			}{status, err}
		}()
	}
	close(start)
	registrations.Wait()
	close(results)
	successes := 0
	for result := range results {
		require.NoError(t, result.err)
		if result.status == 201 {
			successes++
		} else {
			require.Equal(t, 409, result.status)
		}
	}
	require.Equal(t, 1, successes)
	var credentialCount, profileCount int64
	require.NoError(t, database.DB.Table("users").Where("email = ?", "race@example.com").Count(&credentialCount).Error)
	require.NoError(t, identityDB.Table("users").Where("email = ?", "race@example.com").Count(&profileCount).Error)
	require.Equal(t, int64(1), credentialCount)
	require.Equal(t, int64(1), profileCount)
	// Сценарий: источник недоступен — вход по исторической локальной копии запрещён.
	stop()
	login("auth-new", 503)
	login("old", 503)
}

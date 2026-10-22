package account

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	keyrepo "github.com/ivasnev/FinFlow/ff-auth/internal/repository/key_pair"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	tokenservice "github.com/ivasnev/FinFlow/ff-auth/internal/service/token"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type identityStore struct {
	sync.Mutex
	users        map[int64]ffid.UserDTO
	loseResponse bool
	unavailable  bool
}

func (s *identityStore) RegisterUser(_ context.Context, r *ffid.RegisterUserRequest) (*ffid.UserDTO, error) {
	s.Lock()
	defer s.Unlock()
	for _, u := range s.users {
		if strings.EqualFold(u.Email, r.Email) || u.Nickname == r.Nickname {
			return nil, ffid.ErrConflict
		}
	}
	u := ffid.UserDTO{ID: r.UserID, Email: r.Email, Nickname: r.Nickname, Phone: r.Phone}
	s.users[u.ID] = u
	if s.loseResponse {
		s.loseResponse = false
		return nil, ffid.ErrUnavailable
	}
	return &u, nil
}
func (s *identityStore) FindIdentity(_ context.Context, value, kind string) (*ffid.UserDTO, error) {
	s.Lock()
	defer s.Unlock()
	if s.unavailable {
		return nil, ffid.ErrUnavailable
	}
	for _, u := range s.users {
		if (kind == ffid.LookupEmail || kind == ffid.LookupLogin && strings.Contains(value, "@")) && strings.EqualFold(u.Email, value) || (kind == ffid.LookupNickname || kind == ffid.LookupLogin && !strings.Contains(value, "@")) && u.Nickname == value {
			return &u, nil
		}
	}
	return nil, ffid.ErrNotFound
}
func (s *identityStore) GetIdentity(_ context.Context, id int64) (*ffid.UserDTO, error) {
	s.Lock()
	defer s.Unlock()
	if s.unavailable {
		return nil, ffid.ErrUnavailable
	}
	u, ok := s.users[id]
	if !ok {
		return nil, ffid.ErrNotFound
	}
	return &u, nil
}
func (s *identityStore) UpdateIdentity(_ context.Context, id int64, email, nickname *string) (*ffid.UserDTO, error) {
	s.Lock()
	defer s.Unlock()
	u, ok := s.users[id]
	if !ok {
		return nil, ffid.ErrNotFound
	}
	for key, other := range s.users {
		if key != id && email != nil && strings.EqualFold(other.Email, *email) {
			return nil, ffid.ErrConflict
		}
	}
	if email != nil {
		u.Email = *email
	}
	if nickname != nil {
		u.Nickname = *nickname
	}
	s.users[id] = u
	return &u, nil
}

type captchaTransport func(*http.Request) (*http.Response, error)

func (f captchaTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type fixture struct {
	t      *testing.T
	db     *gorm.DB
	m      *Manager
	id     *identityStore
	server *httptest.Server
	tokens service.TokenManager
}

func newFixture(t *testing.T, db *gorm.DB) *fixture {
	require.NoError(t, db.Exec(`TRUNCATE users,account_pending,account_email_tokens,account_captcha,account_limits CASCADE`).Error)
	tokens, err := tokenservice.NewED25519TokenManager(keyrepo.NewKeyPairRepository(db))
	require.NoError(t, err)
	id := &identityStore{users: map[int64]ffid.UserDTO{}}
	cfg := Config{Local: true, Secret: strings.Repeat("test-secret-", 4), AppURL: "http://127.0.0.1:3000", SMTPAddr: "127.0.0.1:1", From: "FinFlow <noreply@finflow.test>", CaptchaProvider: "image", PasswordMin: 8, HashCost: 4, AccessTTL: 15 * time.Minute, RefreshTTL: 24 * time.Hour, MailDaily: 100, MailMonthly: 2000}
	m, err := New(db, cfg, tokens, id)
	require.NoError(t, err)
	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(nil))
	r := router.Group("/api/v1", m.Middleware)
	r.GET("/auth/captcha", m.Captcha)
	r.POST("/auth/register", m.Register)
	r.POST("/auth/login", m.Login)
	r.POST("/auth/refresh", m.Refresh)
	r.POST("/auth/introspect", m.Introspect)
	r.POST("/auth/verification/resend", m.Resend)
	r.POST("/auth/verification/email", m.CorrectEmail)
	r.POST("/auth/verification/confirm", m.Verify)
	r.POST("/auth/verification/logout", m.PendingLogout)
	r.POST("/auth/password/forgot", m.Forgot)
	r.POST("/auth/password/reset", m.Reset)
	r.POST("/auth/password/change", m.ChangePassword)
	r.POST("/auth/email/change", m.ChangeEmail)
	r.POST("/auth/email/confirm", m.ConfirmEmail)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return &fixture{t, db, m, id, server, tokens}
}
func (f *fixture) request(path string, body any, bearer string) (int, map[string]any) {
	f.t.Helper()
	data, err := json.Marshal(body)
	require.NoError(f.t, err)
	req, err := http.NewRequest(http.MethodPost, f.server.URL+"/api/v1"+path, bytes.NewReader(data))
	require.NoError(f.t, err)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(f.t, err)
	defer resp.Body.Close()
	var result map[string]any
	if resp.StatusCode != 204 {
		require.NoError(f.t, json.NewDecoder(resp.Body).Decode(&result))
	}
	return resp.StatusCode, result
}
func (f *fixture) captcha() string {
	f.t.Helper()
	resp, err := http.Get(f.server.URL + "/api/v1/auth/captcha")
	require.NoError(f.t, err)
	defer resp.Body.Close()
	require.Equal(f.t, 200, resp.StatusCode)
	var data struct{ ID, Image string }
	require.NoError(f.t, json.NewDecoder(resp.Body).Decode(&data))
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(data.Image, "data:image/png;base64,"))
	require.NoError(f.t, err)
	picture, err := png.Decode(bytes.NewReader(raw))
	require.NoError(f.t, err)
	patterns := []int{0x3f, 0x06, 0x5b, 0x4f, 0x66, 0x6d, 0x7d, 0x07, 0x7f, 0x6f}
	digits := ""
	for i := 0; i < 6; i++ {
		x, y := 12+i*38, 12
		points := [][2]int{{x + 10, y + 1}, {x + 24, y + 10}, {x + 24, y + 35}, {x + 10, y + 47}, {x + 2, y + 35}, {x + 2, y + 10}, {x + 10, y + 24}}
		bits := 0
		for j, p := range points {
			r, _, _, _ := picture.At(p[0], p[1]).RGBA()
			if r < 30000 {
				bits |= 1 << j
			}
		}
		found := false
		for digit, pattern := range patterns {
			if bits == pattern {
				digits += fmt.Sprint(digit)
				found = true
				break
			}
		}
		require.True(f.t, found)
	}
	return data.ID + ":" + digits
}
func (f *fixture) register(email, nickname string) string {
	f.t.Helper()
	status, data := f.request("/auth/register", map[string]any{"email": email, "nickname": nickname, "password": "password123", "captcha_token": f.captcha()}, "")
	require.Equal(f.t, 202, status, data)
	require.NotContains(f.t, data, "access_token")
	return data["pending_token"].(string)
}
func (f *fixture) mailToken(purpose string) string {
	f.t.Helper()
	var row Outbox
	require.NoError(f.t, f.db.Joins("JOIN account_email_tokens t ON t.id=account_mail_outbox.id").Where("t.purpose=?", purpose).Order("account_mail_outbox.created_at DESC").First(&row).Error)
	aead := f.m.cfg.cipher()
	plain, err := aead.Open(nil, row.Payload[:aead.NonceSize()], row.Payload[aead.NonceSize():], []byte(row.ID.String()))
	require.NoError(f.t, err)
	var d delivery
	require.NoError(f.t, json.Unmarshal(plain, &d))
	parts := strings.Split(d.Body, "#token=")
	require.Len(f.t, parts, 2)
	return strings.Split(parts[1], "\n")[0]
}
func (f *fixture) count(table string) int64 {
	var n int64
	require.NoError(f.t, f.db.Table(table).Count(&n).Error)
	return n
}
func (f *fixture) verify(email, nickname string) map[string]any {
	secret := f.register(email, nickname)
	status, data := f.request("/auth/verification/confirm", map[string]any{"pending_token": secret, "token": f.mailToken("verify")}, "")
	require.Equal(f.t, 200, status, data)
	return data
}
func (f *fixture) clearLimits() { require.NoError(f.t, f.db.Exec(`TRUNCATE account_limits`).Error) }

func TestAccountLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	container, err := postgrescontainer.Run(ctx, "postgres:15-alpine", postgrescontainer.WithDatabase("accounts_test"), postgrescontainer.WithUsername("postgres"), postgrescontainer.WithPassword("postgres"), testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(45*time.Second)))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, container.Terminate(ctx)) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	sqlDB.SetMaxOpenConns(20)
	files, err := filepath.Glob("../repository/migrations/*.up.sql")
	require.NoError(t, err)
	require.NotEmpty(t, files)
	for _, file := range files {
		source, err := os.ReadFile(file)
		require.NoError(t, err)
		require.NoError(t, db.Exec(string(source)).Error, file)
	}
	t.Run("pending has no application credentials and requires matching browser", func(t *testing.T) {
		f := newFixture(t, db)
		secret := f.register("person@example.test", "person")
		require.Zero(t, f.count("users"))
		require.Zero(t, f.count("sessions"))
		emailToken := f.mailToken("verify")
		status, _ := f.request("/auth/verification/confirm", map[string]any{"pending_token": randomToken(), "token": emailToken}, "")
		require.Equal(t, 401, status)
		status, _ = f.request("/auth/introspect", nil, secret)
		require.Equal(t, 401, status)
		status, data := f.request("/auth/verification/confirm", map[string]any{"pending_token": secret, "token": emailToken}, "")
		require.Equal(t, 200, status, data)
		require.EqualValues(t, 1, f.count("users"))
		require.Zero(t, f.count("account_pending"))
		status, _ = f.request("/auth/introspect", nil, data["access_token"].(string))
		require.Equal(t, 200, status)
		status, _ = f.request("/auth/verification/confirm", map[string]any{"pending_token": secret, "token": emailToken}, "")
		require.NotEqual(t, 200, status)
	})
	t.Run("CAPTCHA cannot be skipped or replayed", func(t *testing.T) {
		f := newFixture(t, db)
		body := map[string]any{"email": "a@example.test", "nickname": "aaaa", "password": "password123"}
		status, _ := f.request("/auth/register", body, "")
		require.Equal(t, 400, status)
		body["captcha_token"] = f.captcha()
		status, _ = f.request("/auth/register", body, "")
		require.Equal(t, 202, status)
		body["email"] = "b@example.test"
		body["nickname"] = "bbbb"
		status, _ = f.request("/auth/register", body, "")
		require.Equal(t, 400, status)
		require.EqualValues(t, 1, f.count("account_pending"))
	})
	t.Run("email correction invalidates old link and needs password", func(t *testing.T) {
		f := newFixture(t, db)
		secret := f.register("typo@example.test", "person")
		old := f.mailToken("verify")
		f.clearLimits()
		status, _ := f.request("/auth/verification/email", map[string]any{"pending_token": secret, "email": "correct@example.test", "password": "wrong"}, "")
		require.Equal(t, 400, status)
		status, data := f.request("/auth/verification/email", map[string]any{"pending_token": secret, "email": "correct@example.test", "password": "password123"}, "")
		require.Equal(t, 202, status, data)
		current := f.mailToken("verify")
		require.NotEqual(t, old, current)
		status, _ = f.request("/auth/verification/confirm", map[string]any{"pending_token": secret, "token": old}, "")
		require.Equal(t, 400, status)
		status, data = f.request("/auth/verification/confirm", map[string]any{"pending_token": secret, "token": current}, "")
		require.Equal(t, 200, status, data)
	})
	t.Run("resend is throttled and rotates tokens", func(t *testing.T) {
		f := newFixture(t, db)
		secret := f.register("a@example.test", "person")
		old := f.mailToken("verify")
		status, _ := f.request("/auth/verification/resend", map[string]any{"pending_token": secret}, "")
		require.Equal(t, 429, status)
		f.clearLimits()
		status, _ = f.request("/auth/verification/resend", map[string]any{"pending_token": secret}, "")
		require.Equal(t, 200, status)
		require.NotEqual(t, old, f.mailToken("verify"))
		require.EqualValues(t, 1, f.count("account_email_tokens"))
	})
	t.Run("lost identity response reuses reserved ID", func(t *testing.T) {
		f := newFixture(t, db)
		f.id.loseResponse = true
		_ = f.verify("a@example.test", "person")
		require.EqualValues(t, 1, f.count("users"))
		require.Len(t, f.id.users, 1)
	})
	t.Run("expired signup releases address and is cleaned", func(t *testing.T) {
		f := newFixture(t, db)
		secret := f.register("a@example.test", "person")
		raw := f.mailToken("verify")
		require.NoError(t, db.Exec(`UPDATE account_pending SET expires_at=now()-interval '1 second'`).Error)
		require.NoError(t, f.m.Cleanup(ctx))
		require.Zero(t, f.count("account_pending"))
		require.Zero(t, f.count("account_mail_outbox"))
		status, _ := f.request("/auth/verification/confirm", map[string]any{"pending_token": secret, "token": raw}, "")
		require.NotEqual(t, 200, status)
		f.clearLimits()
		_ = f.register("a@example.test", "person")
	})
	t.Run("old accounts need verification and preserve their data", func(t *testing.T) {
		f := newFixture(t, db)
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 4)
		u := credentials{Email: "old@example.test", Nickname: "person", PasswordHash: string(hash), CreatedAt: time.Now(), UpdatedAt: time.Now()}
		require.NoError(t, db.Create(&u).Error)
		f.id.users[u.ID] = ffid.UserDTO{ID: u.ID, Email: u.Email, Nickname: u.Nickname}
		status, data := f.request("/auth/login", map[string]any{"login": u.Email, "password": "password123"}, "")
		require.Equal(t, 202, status, data)
		require.NotContains(t, data, "access_token")
		status, _ = f.request("/auth/verification/resend", map[string]any{"pending_token": data["pending_token"]}, "")
		require.Equal(t, 200, status)
		status, data = f.request("/auth/verification/confirm", map[string]any{"pending_token": data["pending_token"], "token": f.mailToken("verify")}, "")
		require.Equal(t, 200, status, data)
		require.EqualValues(t, 1, f.count("users"))
	})
	t.Run("reset revokes access refresh and email changes with generic request responses", func(t *testing.T) {
		f := newFixture(t, db)
		session := f.verify("a@example.test", "person")
		access := session["access_token"].(string)
		f.clearLimits()
		status, _ := f.request("/auth/email/change", map[string]any{"email": "new@example.test", "current_password": "password123"}, access)
		require.Equal(t, 200, status)
		changeToken := f.mailToken("email")
		f.clearLimits()
		status, known := f.request("/auth/password/forgot", map[string]any{"email": "a@example.test", "captcha_token": f.captcha()}, "")
		require.Equal(t, 200, status)
		reset := f.mailToken("reset")
		status, unknown := f.request("/auth/password/forgot", map[string]any{"email": "unknown@example.test", "captcha_token": f.captcha()}, "")
		require.Equal(t, 200, status)
		require.Equal(t, known, unknown)
		status, _ = f.request("/auth/password/reset", map[string]any{"token": reset, "password": "new-password123"}, "")
		require.Equal(t, 200, status)
		status, _ = f.request("/auth/introspect", nil, access)
		require.Equal(t, 401, status)
		status, _ = f.request("/auth/refresh", map[string]any{"refresh_token": session["refresh_token"]}, "")
		require.Equal(t, 401, status)
		status, _ = f.request("/auth/email/confirm", map[string]any{"token": changeToken}, access)
		require.Equal(t, 401, status)
		status, _ = f.request("/auth/password/reset", map[string]any{"token": reset, "password": "new-password123"}, "")
		require.Equal(t, 400, status)
		status, _ = f.request("/auth/login", map[string]any{"login": "a@example.test", "password": "password123"}, "")
		require.Equal(t, 401, status)
		status, _ = f.request("/auth/login", map[string]any{"login": "a@example.test", "password": "new-password123"}, "")
		require.Equal(t, 200, status)
	})
	t.Run("email change keeps old email until confirmation then revokes all sessions", func(t *testing.T) {
		f := newFixture(t, db)
		session := f.verify("a@example.test", "person")
		access := session["access_token"].(string)
		f.clearLimits()
		status, _ := f.request("/auth/email/change", map[string]any{"email": "new@example.test", "current_password": "bad"}, access)
		require.Equal(t, 400, status)
		status, _ = f.request("/auth/email/change", map[string]any{"email": "new@example.test", "current_password": "password123"}, access)
		require.Equal(t, 200, status)
		raw := f.mailToken("email")
		u, _ := f.id.FindIdentity(ctx, "a@example.test", ffid.LookupEmail)
		require.NotNil(t, u)
		status, _ = f.request("/auth/email/confirm", map[string]any{"token": raw}, access)
		require.Equal(t, 200, status)
		_, err := f.id.FindIdentity(ctx, "a@example.test", ffid.LookupEmail)
		require.ErrorIs(t, err, ffid.ErrNotFound)
		status, _ = f.request("/auth/introspect", nil, access)
		require.Equal(t, 401, status)
		status, _ = f.request("/auth/login", map[string]any{"login": "new@example.test", "password": "password123"}, "")
		require.Equal(t, 200, status)
	})
	t.Run("atomic global quota survives concurrent requests", func(t *testing.T) {
		f := newFixture(t, db)
		var accepted atomic.Int32
		var wg sync.WaitGroup
		for range 40 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := f.m.limit(db, "shared", 5, time.Hour)
				if err == nil {
					accepted.Add(1)
				} else {
					require.ErrorIs(t, err, errLimited)
				}
			}()
		}
		wg.Wait()
		require.EqualValues(t, 5, accepted.Load())
	})
	t.Run("mail global quota cannot be bypassed with new identities", func(t *testing.T) {
		f := newFixture(t, db)
		f.m.cfg.MailDaily = 1
		_ = f.register("a@example.test", "aaaa")
		status, _ := f.request("/auth/register", map[string]any{"email": "b@example.test", "nickname": "bbbb", "password": "password123", "captcha_token": f.captcha()}, "")
		require.Equal(t, 429, status)
		require.EqualValues(t, 1, f.count("account_pending"))
		require.EqualValues(t, 1, f.count("account_mail_outbox"))
	})
	t.Run("local mode never accepts remote SMTP or production CAPTCHA bypass", func(t *testing.T) {
		f := newFixture(t, db)
		cfg := f.m.cfg
		cfg.Local = false
		_, err := New(db, cfg, f.tokens, f.id)
		require.Error(t, err)
		cfg = f.m.cfg
		cfg.SMTPAddr = "example.com:25"
		_, err = New(db, cfg, f.tokens, f.id)
		require.Error(t, err)
	})
	t.Run("introspection fails closed and middleware preserves unavailable status", func(t *testing.T) {
		f := newFixture(t, db)
		session := f.verify("a@example.test", "person")
		client := auth.NewClient(f.server.URL, time.Minute)
		_, err := client.ValidateToken(session["access_token"].(string))
		require.Error(t, err) /* server below supplies the public key */
		f.server.Config.Handler.(*gin.Engine).GET("/api/v1/auth/public-key", func(c *gin.Context) { c.String(200, base64.StdEncoding.EncodeToString(f.tokens.GetPublicKey())) })
		_, err = client.ValidateToken(session["access_token"].(string))
		require.NoError(t, err)
		require.NoError(t, db.Exec(`DELETE FROM sessions`).Error)
		_, err = client.ValidateToken(session["access_token"].(string))
		require.Error(t, err)
		f.server.Close()
		_, err = client.ValidateToken(session["access_token"].(string))
		require.ErrorIs(t, err, auth.ErrUnavailable)
	})
	t.Run("expired outbox is discarded without any SMTP attempt", func(t *testing.T) {
		f := newFixture(t, db)
		_ = f.register("a@example.test", "person")
		before := f.count("account_limits")
		require.NoError(t, db.Exec(`UPDATE account_email_tokens SET expires_at=now()-interval '1 second'`).Error)
		require.NoError(t, f.m.DeliverOne(ctx))
		require.Zero(t, f.count("account_mail_outbox"))
		require.Equal(t, before, f.count("account_limits"))
	})
	t.Run("four slow bodies do not occupy CPU admission slots", func(t *testing.T) {
		f := newFixture(t, db)
		_ = f.verify("a@example.test", "person")
		var connections []net.Conn
		for range 4 {
			conn, err := net.Dial("tcp", strings.TrimPrefix(f.server.URL, "http://"))
			require.NoError(t, err)
			connections = append(connections, conn)
			_, err = fmt.Fprint(conn, "POST /api/v1/auth/login HTTP/1.1\r\nHost: localhost\r\nContent-Type: application/json\r\nContent-Length: 10000\r\n\r\n{")
			require.NoError(t, err)
		}
		defer func() {
			for _, conn := range connections {
				_ = conn.Close()
			}
		}()
		status, data := f.request("/auth/login", map[string]any{"login": "a@example.test", "password": "password123"}, "")
		require.Equal(t, 200, status, data)
	})
	t.Run("provision retries after identity outage without allocating a second user", func(t *testing.T) {
		f := newFixture(t, db)
		secret := f.register("a@example.test", "person")
		raw := f.mailToken("verify")
		f.id.unavailable = true
		status, _ := f.request("/auth/verification/confirm", map[string]any{"pending_token": secret, "token": raw}, "")
		require.Equal(t, 503, status)
		require.EqualValues(t, 1, f.count("users"))
		f.id.unavailable = false
		status, data := f.request("/auth/verification/confirm", map[string]any{"pending_token": secret, "token": raw}, "")
		require.Equal(t, 200, status, data)
		require.EqualValues(t, 1, f.count("users"))
		require.Len(t, f.id.users, 1)
	})
	t.Run("concurrent confirmations create one account and one session", func(t *testing.T) {
		f := newFixture(t, db)
		secret := f.register("a@example.test", "person")
		raw := f.mailToken("verify")
		var wg sync.WaitGroup
		statuses := make(chan int, 2)
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				status, _ := f.request("/auth/verification/confirm", map[string]any{"pending_token": secret, "token": raw}, "")
				statuses <- status
			}()
		}
		wg.Wait()
		close(statuses)
		success := 0
		for status := range statuses {
			if status == 200 {
				success++
			} else {
				require.Contains(t, []int{400, 401}, status)
			}
		}
		require.Equal(t, 1, success)
		require.EqualValues(t, 1, f.count("users"))
		require.EqualValues(t, 1, f.count("sessions"))
	})
	t.Run("reset cannot race old-password login or refresh into a surviving session", func(t *testing.T) {
		f := newFixture(t, db)
		session := f.verify("a@example.test", "person")
		f.clearLimits()
		status, _ := f.request("/auth/password/forgot", map[string]any{"email": "a@example.test", "captcha_token": f.captcha()}, "")
		require.Equal(t, 200, status)
		reset := f.mailToken("reset")
		lock := db.Begin()
		require.NoError(t, lock.Exec(`SELECT id FROM users FOR UPDATE`).Error)
		type result struct {
			kind   string
			status int
			data   map[string]any
		}
		results := make(chan result, 3)
		var wg sync.WaitGroup
		operations := []struct {
			kind, path string
			body       map[string]any
		}{{"reset", "/auth/password/reset", map[string]any{"token": reset, "password": "new-password123"}}, {"refresh", "/auth/refresh", map[string]any{"refresh_token": session["refresh_token"]}}, {"login", "/auth/login", map[string]any{"login": "a@example.test", "password": "password123"}}}
		for _, op := range operations {
			wg.Add(1)
			go func() {
				defer wg.Done()
				status, data := f.request(op.path, op.body, "")
				results <- result{op.kind, status, data}
			}()
		}
		time.Sleep(20 * time.Millisecond)
		require.NoError(t, lock.Commit().Error)
		wg.Wait()
		close(results)
		for r := range results {
			if r.kind == "reset" {
				require.Equal(t, 200, r.status, r.data)
			} else {
				require.Contains(t, []int{200, 401}, r.status)
				if r.status == 200 {
					status, _ = f.request("/auth/introspect", nil, r.data["access_token"].(string))
					require.Equal(t, 401, status)
				}
			}
		}
		require.Zero(t, f.count("sessions"))
	})
	t.Run("phone is preserved until verified provisioning", func(t *testing.T) {
		f := newFixture(t, db)
		status, data := f.request("/auth/register", map[string]any{"email": "a@example.test", "nickname": "person", "password": "password123", "phone": "+70000000000", "captcha_token": f.captcha()}, "")
		require.Equal(t, 202, status, data)
		status, data = f.request("/auth/verification/confirm", map[string]any{"pending_token": data["pending_token"], "token": f.mailToken("verify")}, "")
		require.Equal(t, 200, status, data)
		identity, err := f.id.FindIdentity(ctx, "a@example.test", ffid.LookupEmail)
		require.NoError(t, err)
		require.NotNil(t, identity.Phone)
		require.Equal(t, "+70000000000", *identity.Phone)
	})
	t.Run("manual logins record history while refresh does not", func(t *testing.T) {
		f := newFixture(t, db)
		_ = f.verify("a@example.test", "person")
		before := f.count("login_history")
		status, data := f.request("/auth/login", map[string]any{"login": "a@example.test", "password": "password123"}, "")
		require.Equal(t, 200, status, data)
		require.Equal(t, before+1, f.count("login_history"))
		require.EqualValues(t, 1, f.count("devices"))
		status, _ = f.request("/auth/refresh", map[string]any{"refresh_token": data["refresh_token"]}, "")
		require.Equal(t, 200, status)
		require.Equal(t, before+1, f.count("login_history"))
	})

	t.Run("SmartCaptcha uses official endpoint and fails closed on errors and host mismatch", func(t *testing.T) {
		f := newFixture(t, db)
		f.m.cfg.CaptchaProvider = "smartcaptcha"
		f.m.cfg.CaptchaSecret = "test-server-key"
		calls := 0
		f.m.http.Transport = captchaTransport(func(r *http.Request) (*http.Response, error) {
			calls++
			require.Equal(t, "https://smartcaptcha.cloud.yandex.ru/validate", r.URL.String())
			require.NoError(t, r.ParseForm())
			require.Equal(t, "test-server-key", r.Form.Get("secret"))
			require.Equal(t, "127.0.0.1", r.Form.Get("ip"))
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"status":"ok","host":"127.0.0.1:3000"}`)), Header: make(http.Header)}, nil
		})
		require.NoError(t, f.m.captcha(ctx, "first-token", "127.0.0.1"))
		require.Error(t, f.m.captcha(ctx, "first-token", "127.0.0.1"))
		require.Equal(t, 1, calls)
		f.m.http.Transport = captchaTransport(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"status":"ok","host":"evil.example"}`)), Header: make(http.Header)}, nil
		})
		require.Error(t, f.m.captcha(ctx, "second-token", "127.0.0.1"))
		f.m.http.Transport = captchaTransport(func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 503, Body: io.NopCloser(strings.NewReader("unavailable")), Header: make(http.Header)}, nil
		})
		require.Error(t, f.m.captcha(ctx, "third-token", "127.0.0.1"))
	})

	t.Run("outbox retries and clears encrypted delivery payload", func(t *testing.T) {
		f := newFixture(t, db)
		_ = f.register("a@example.test", "person")
		err := f.m.DeliverOne(ctx)
		require.Error(t, err)
		var row Outbox
		require.NoError(t, db.First(&row).Error)
		require.Equal(t, 1, row.Attempts)
		require.NotEmpty(t, row.Payload)
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer listener.Close()
		delivered := make(chan string, 1)
		go func() {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			defer conn.Close()
			reader := bufio.NewReader(conn)
			fmt.Fprint(conn, "220 local SMTP\r\n")
			var body strings.Builder
			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				switch {
				case strings.HasPrefix(line, "EHLO"):
					fmt.Fprint(conn, "250 localhost\r\n")
				case strings.HasPrefix(line, "DATA"):
					fmt.Fprint(conn, "354 go\r\n")
					for {
						line, err = reader.ReadString('\n')
						if err != nil {
							return
						}
						if line == ".\r\n" {
							break
						}
						body.WriteString(line)
					}
					fmt.Fprint(conn, "250 saved\r\n")
				case strings.HasPrefix(line, "QUIT"):
					fmt.Fprint(conn, "221 bye\r\n")
					delivered <- body.String()
					return
				default:
					fmt.Fprint(conn, "250 ok\r\n")
				}
			}
		}()
		f.m.cfg.SMTPAddr = listener.Addr().String()
		require.NoError(t, db.Model(&row).Updates(map[string]any{"available_at": time.Now().Add(-time.Second)}).Error)
		require.NoError(t, f.m.DeliverOne(ctx))
		require.Contains(t, <-delivered, "#token=")
		require.NoError(t, db.First(&row).Error)
		require.NotNil(t, row.SentAt)
		require.Empty(t, row.Payload)
		require.NoError(t, f.m.DeliverOne(ctx))
		require.EqualValues(t, 1, f.count("account_mail_outbox"))
	})
}

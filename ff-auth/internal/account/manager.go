package account

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errCurrentPassword = errors.New("Текущий пароль указан неверно.")
var errCaptcha = errors.New("Проверка капчи не пройдена. Обновите её и попробуйте ещё раз.")
var errInvalid = errors.New("ссылка недействительна или срок её действия истёк")
var errLimited = errors.New("слишком много попыток, попробуйте позже")
var errUnauthorized = errors.New("войдите в аккаунт заново")

type Manager struct {
	db       *gorm.DB
	cfg      Config
	tokens   service.TokenManager
	identity service.Identity
	http     *http.Client
	slots    chan struct{}
}

func New(db *gorm.DB, cfg Config, tokens service.TokenManager, identity service.Identity) (*Manager, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &Manager{db: db, cfg: cfg, tokens: tokens, identity: identity, http: &http.Client{Timeout: 5 * time.Second}, slots: make(chan struct{}, 4)}, nil
}

func randomToken() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b[:])
}
func digest(s string) string { v := sha256.Sum256([]byte(s)); return hex.EncodeToString(v[:]) }
func (m *Manager) key(s string) string {
	h := hmac.New(sha256.New, []byte(m.cfg.Secret))
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func (m *Manager) answer(c *gin.Context, err error) {
	status := http.StatusServiceUnavailable
	msg := "Сервис временно недоступен. Попробуйте позже."
	switch {
	case errors.Is(err, errCurrentPassword):
		status = 400
		msg = errCurrentPassword.Error()
	case errors.Is(err, errCaptcha):
		status = 400
		msg = errCaptcha.Error()
	case errors.Is(err, errInvalid):
		status = 400
		msg = errInvalid.Error()
	case errors.Is(err, errUnauthorized):
		status = 401
		msg = errUnauthorized.Error()
	case errors.Is(err, errLimited):
		status = 429
		msg = errLimited.Error()
		c.Header("Retry-After", "60")
	}
	c.JSON(status, gin.H{"error": msg})
}

func bind(c *gin.Context, v any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 16<<10)
	if err := c.ShouldBindJSON(v); err != nil {
		c.JSON(400, gin.H{"error": "Проверьте заполнение полей."})
		return false
	}
	return true
}
func bearer(c *gin.Context) string {
	s := strings.Split(c.GetHeader("Authorization"), " ")
	if len(s) == 2 && s[0] == "Bearer" && len(s[1]) <= 8192 {
		return s[1]
	}
	return ""
}
func (m *Manager) pending(tx *gorm.DB, secret string) (*Pending, error) {
	if len(secret) != 43 {
		return nil, errUnauthorized
	}
	var p Pending
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("secret_hash = ? AND expires_at > now()", digest(secret)).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errUnauthorized
	}
	return &p, err
}
func pendingResponse(c *gin.Context, p *Pending, secret string) {
	c.JSON(202, gin.H{"verification_required": true, "pending_token": secret, "email": p.Email, "expires_at": p.ExpiresAt})
}

// ValidateAccess checks a cryptographically valid token against a live session. No positive cache: revocation is immediate.
func (m *Manager) ValidateAccess(ctx context.Context, token string) (*service.TokenPayload, error) {
	if token == "" || len(token) > 8192 {
		return nil, errUnauthorized
	}
	payload, err := m.tokens.ValidateToken(token)
	if err != nil {
		return nil, errUnauthorized
	}
	var count int64
	err = m.db.WithContext(ctx).Raw(`SELECT count(*) FROM account_access a JOIN sessions s ON s.id=a.session_id JOIN account_security v ON v.user_id=s.user_id WHERE a.token_hash=? AND a.expires_at>now() AND s.expires_at>now() AND s.user_id=?`, digest(token), payload.UserID).Scan(&count).Error
	if err != nil {
		return nil, err
	}
	if count != 1 {
		return nil, errUnauthorized
	}
	return payload, nil
}
func (m *Manager) current(c *gin.Context) (*service.TokenPayload, bool) {
	p, err := m.ValidateAccess(c.Request.Context(), bearer(c))
	if err != nil {
		m.answer(c, err)
		return nil, false
	}
	return p, true
}
func (m *Manager) Introspect(c *gin.Context) {
	p, ok := m.current(c)
	if !ok {
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(200, gin.H{"active": true, "user_id": p.UserID})
}

// Route middleware is deliberately installed before generated auth routes; requests and capability URLs are never logged here.
func (m *Manager) ConfigHTTP(c *gin.Context) {
	c.JSON(200, gin.H{"captcha_provider": m.cfg.CaptchaProvider, "captcha_site_key": m.cfg.CaptchaSiteKey, "local_mode": m.cfg.Local})
}

func (m *Manager) Middleware(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	if strings.HasPrefix(c.Request.URL.Path, "/api/v1/auth/") && c.Request.URL.Path != "/api/v1/auth/introspect" {
		if err := m.admit(c, "request", 120, 3000, time.Minute, ""); err != nil {
			m.answer(c, err)
			c.Abort()
			return
		}
	}
	c.Next()
}

func IsUnauthorized(err error) bool { return errors.Is(err, errUnauthorized) }

// WorkSlot is acquired only after parsing a bounded body; slow senders cannot exhaust the CPU slots.
func (m *Manager) workSlot(c *gin.Context) (func(), bool) {
	select {
	case m.slots <- struct{}{}:
		return func() { <-m.slots }, true
	default:
		m.answer(c, errLimited)
		return nil, false
	}
}

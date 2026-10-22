package account

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service/password"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func canonicalEmail(value string) (string, error) {
	value = strings.TrimSpace(value)
	a, err := mail.ParseAddress(value)
	if err != nil || a.Address != value || len(value) > 254 || strings.ContainsAny(value, "\r\n") {
		return "", errInvalid
	}
	return strings.ToLower(value), nil
}
func (m *Manager) lockedUser(tx *gorm.DB, id int64) (*credentials, error) {
	var u credentials
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errUnauthorized
	}
	return &u, err
}
func (m *Manager) verified(tx *gorm.DB, id int64, email string) (bool, error) {
	var n int64
	err := tx.Model(&Security{}).Where("user_id=? AND lower(verified_email)=lower(?)", id, email).Count(&n).Error
	return n == 1, err
}
func (m *Manager) session(tx *gorm.DB, u *credentials, identity *ffid.UserDTO, ip string) (*service.AccessDataParams, error) {
	var roles []string
	if err := tx.Table("roles").Select("roles.name").Joins("JOIN user_roles ON user_roles.role_id=roles.id").Where("user_roles.user_id=?", u.ID).Scan(&roles).Error; err != nil {
		return nil, err
	}
	access, refresh, expires, refreshExpires, err := m.tokens.GenerateTokenPair(u.ID, roles, m.cfg.AccessTTL, m.cfg.RefreshTTL)
	if err != nil {
		return nil, err
	}
	sid := uuid.New()
	if err := tx.Exec(`INSERT INTO sessions(id,user_id,refresh_token,ip_address,expires_at) VALUES(?,?,?,NULLIF(?,'')::inet,?)`, sid, u.ID, refresh, ip, time.Unix(refreshExpires, 0)).Error; err != nil {
		return nil, err
	}
	if err := tx.Exec(`INSERT INTO account_access(token_hash,session_id,expires_at) VALUES(?,?,?)`, digest(access), sid, time.Unix(expires, 0)).Error; err != nil {
		return nil, err
	}
	return &service.AccessDataParams{AccessToken: access, RefreshToken: refresh, ExpiresAt: time.Unix(expires, 0), User: service.ShortUserParams{Id: u.ID, Email: identity.Email, Nickname: identity.Nickname, Roles: roles}}, nil
}
func authResponse(c *gin.Context, v *service.AccessDataParams) {
	c.JSON(200, gin.H{"access_token": v.AccessToken, "refresh_token": v.RefreshToken, "expires_at": v.ExpiresAt, "user": gin.H{"id": v.User.Id, "email": v.User.Email, "nickname": v.User.Nickname, "roles": v.User.Roles}})
}
func (m *Manager) newPending(tx *gorm.DB, email, nickname string, name *string, hash string, id *int64) (*Pending, string, error) {
	secret := randomToken()
	p := Pending{ID: uuid.New(), SecretHash: digest(secret), Email: email, OriginalEmail: email, Nickname: nickname, Name: name, PasswordHash: hash, UserID: id, ExistingUser: id != nil, ExpiresAt: time.Now().Add(24 * time.Hour), CreatedAt: time.Now()}
	if err := tx.Create(&p).Error; err != nil {
		return nil, "", err
	}
	return &p, secret, nil
}

func (m *Manager) Register(c *gin.Context) {
	var req struct {
		Email        string  `json:"email"`
		Nickname     string  `json:"nickname"`
		Password     string  `json:"password"`
		Name         *string `json:"name"`
		Phone        *string `json:"phone"`
		CaptchaToken string  `json:"captcha_token"`
	}
	if !bind(c, &req) {
		return
	}
	release, ok := m.workSlot(c)
	if !ok {
		return
	}
	defer release()
	email, err := canonicalEmail(req.Email)
	if err != nil {
		m.answer(c, errInvalid)
		return
	}
	if err = m.admit(c, "register", 5, 100, time.Hour, email); err != nil {
		m.answer(c, err)
		return
	}
	if len(req.Nickname) < 3 || len(req.Nickname) > 50 || strings.TrimSpace(req.Nickname) != req.Nickname || strings.ContainsAny(req.Nickname, "@\r\n") {
		c.JSON(400, gin.H{"error": "Никнейм: от 3 до 50 символов, без @ и пробелов по краям."})
		return
	}
	if err = password.Validate(req.Password, m.cfg.PasswordMin); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if err = m.captcha(c.Request.Context(), req.CaptchaToken, c.ClientIP()); err != nil {
		m.answer(c, err)
		return
	}
	for _, q := range []struct{ value, kind string }{{email, ffid.LookupEmail}, {req.Nickname, ffid.LookupNickname}} {
		_, err = m.identity.FindIdentity(c.Request.Context(), q.value, q.kind)
		if err == nil {
			c.JSON(409, gin.H{"error": "Email или никнейм уже используется. Попробуйте войти или восстановить пароль."})
			return
		}
		if !errors.Is(err, ffid.ErrNotFound) {
			m.answer(c, err)
			return
		}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), m.cfg.HashCost)
	if err != nil {
		m.answer(c, err)
		return
	}
	var p *Pending
	var secret string
	err = m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		p, secret, err = m.newPending(tx, email, req.Nickname, req.Name, string(hash), nil)
		if err != nil {
			return err
		}
		if req.Phone != nil {
			value := strings.TrimSpace(*req.Phone)
			if value != "" {
				if len(value) > 32 {
					return errInvalid
				}
				p.Phone = &value
				if err := tx.Model(p).Update("phone", value).Error; err != nil {
					return err
				}
			}
		}
		return m.enqueue(tx, "verify", p, nil, email, string(hash), c.ClientIP())
	})
	if err != nil {
		m.answer(c, err)
		return
	}
	pendingResponse(c, p, secret)
}

func (m *Manager) Login(c *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if !bind(c, &req) {
		return
	}
	release, ok := m.workSlot(c)
	if !ok {
		return
	}
	defer release()
	login := strings.TrimSpace(req.Login)
	if strings.Contains(login, "@") {
		login = strings.ToLower(login)
	}
	if err := m.admit(c, "login", 10, 500, 15*time.Minute, strings.ToLower(login)); err != nil {
		m.answer(c, err)
		return
	}
	identity, err := m.identity.FindIdentity(c.Request.Context(), login, ffid.LookupLogin)
	if errors.Is(err, ffid.ErrNotFound) {
		// Pending sign-ups can be resumed in a different browser only with their password.
		var p Pending
		find := m.db.WithContext(c.Request.Context()).Where("(lower(email)=? OR nickname=?) AND expires_at>now() AND existing_user=false", strings.ToLower(login), login).Order("created_at DESC").First(&p).Error
		if find == nil && bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(req.Password)) == nil {
			secret := randomToken()
			err = m.db.WithContext(c.Request.Context()).Model(&Pending{}).Where("id=? AND password_hash=? AND expires_at>now()", p.ID, p.PasswordHash).Update("secret_hash", digest(secret)).Error
			if err != nil {
				m.answer(c, err)
				return
			}
			pendingResponse(c, &p, secret)
			return
		}
		c.JSON(401, gin.H{"error": "Неверный логин или пароль."})
		return
	}
	if err != nil {
		m.answer(c, err)
		return
	}
	var response *service.AccessDataParams
	var p *Pending
	var secret string
	err = m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		u, err := m.lockedUser(tx, identity.ID)
		if err != nil {
			return err
		}
		if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)) != nil {
			return errUnauthorized
		}
		yes, err := m.verified(tx, u.ID, identity.Email)
		if err != nil {
			return err
		}
		if !yes {
			p, secret, err = m.newPending(tx, strings.ToLower(identity.Email), identity.Nickname, nil, u.PasswordHash, &u.ID)
			return err
		}
		response, err = m.session(tx, u, identity, c.ClientIP())
		if err != nil {
			return err
		}
		return m.recordLogin(tx, u.ID, c.ClientIP(), c.GetHeader("User-Agent"))
	})
	if err != nil {
		m.answer(c, err)
		return
	}
	if p != nil {
		pendingResponse(c, p, secret)
		return
	}
	authResponse(c, response)
}

func (m *Manager) Refresh(c *gin.Context) {
	var req struct {
		Token string `json:"refresh_token"`
	}
	if !bind(c, &req) {
		return
	}
	if len(req.Token) > 8192 || req.Token == "" {
		m.answer(c, errUnauthorized)
		return
	}
	if err := m.admit(c, "refresh", 120, 3000, time.Minute, ""); err != nil {
		m.answer(c, err)
		return
	}
	var row struct {
		ID     uuid.UUID
		UserID int64
	}
	err := m.db.WithContext(c.Request.Context()).Table("sessions").Select("id,user_id").Where("refresh_token=? AND expires_at>now()", req.Token).Take(&row).Error
	if err != nil {
		m.answer(c, errUnauthorized)
		return
	}
	identity, err := m.identity.GetIdentity(c.Request.Context(), row.UserID)
	if err != nil {
		m.answer(c, err)
		return
	}
	var response *service.AccessDataParams
	var p *Pending
	var secret string
	err = m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		u, err := m.lockedUser(tx, row.UserID)
		if err != nil {
			return err
		}
		// User lock serializes refresh/login against password reset and email changes.
		result := tx.Exec(`DELETE FROM sessions WHERE id=? AND refresh_token=? AND expires_at>now()`, row.ID, req.Token)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errUnauthorized
		}
		yes, err := m.verified(tx, u.ID, identity.Email)
		if err != nil {
			return err
		}
		if !yes {
			p, secret, err = m.newPending(tx, strings.ToLower(identity.Email), identity.Nickname, nil, u.PasswordHash, &u.ID)
			return err
		}
		response, err = m.session(tx, u, identity, c.ClientIP())
		return err
	})
	if err != nil {
		m.answer(c, err)
		return
	}
	if p != nil {
		pendingResponse(c, p, secret)
		return
	}
	authResponse(c, response)
}

func (m *Manager) recordLogin(tx *gorm.DB, userID int64, ip, userAgent string) error {
	if len(userAgent) > 1024 {
		userAgent = userAgent[:1024]
	}
	device := m.key(fmt.Sprintf("device:%d:%s:%s", userID, ip, userAgent))
	if err := tx.Exec(`INSERT INTO devices(user_id,device_id,user_agent,last_login) VALUES(?,?,?,now()) ON CONFLICT(device_id) DO UPDATE SET last_login=now()`, userID, device, userAgent).Error; err != nil {
		return err
	}
	return tx.Exec(`INSERT INTO login_history(user_id,ip_address,user_agent) VALUES(?,?::inet,?)`, userID, ip, userAgent).Error
}

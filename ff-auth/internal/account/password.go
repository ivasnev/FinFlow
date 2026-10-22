package account

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service/password"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (m *Manager) Forgot(c *gin.Context) {
	var req struct {
		Email        string `json:"email"`
		CaptchaToken string `json:"captcha_token"`
	}
	if !bind(c, &req) {
		return
	}
	email, err := canonicalEmail(req.Email)
	if err != nil {
		m.answer(c, errInvalid)
		return
	}
	if err = m.admit(c, "forgot", 5, 100, time.Hour, email); err != nil {
		m.answer(c, err)
		return
	}
	if err = m.captcha(c.Request.Context(), req.CaptchaToken, c.ClientIP()); err != nil {
		m.answer(c, err)
		return
	}
	identity, err := m.identity.FindIdentity(c.Request.Context(), email, ffid.LookupEmail)
	if err != nil && !errors.Is(err, ffid.ErrNotFound) {
		m.answer(c, err)
		return
	}
	if err == nil {
		err = m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
			u, err := m.lockedUser(tx, identity.ID)
			if err != nil {
				return err
			}
			return m.enqueue(tx, "reset", nil, &u.ID, strings.ToLower(identity.Email), u.PasswordHash, c.ClientIP())
		})
		if err != nil && !errors.Is(err, errLimited) {
			m.answer(c, err)
			return
		}
	}
	c.JSON(200, gin.H{"message": "Если аккаунт с таким адресом существует, мы отправим письмо для восстановления пароля."})
}

func (m *Manager) invalidate(tx *gorm.DB, id int64) error {
	if err := tx.Exec(`DELETE FROM sessions WHERE user_id=?`, id).Error; err != nil {
		return err
	}
	if err := tx.Where("user_id=?", id).Delete(&Pending{}).Error; err != nil {
		return err
	}
	return tx.Where("user_id=?", id).Delete(&EmailToken{}).Error
}
func (m *Manager) Reset(c *gin.Context) {
	var req struct {
		Token    string `json:"token"`
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
	if err := m.admit(c, "reset", 10, 100, time.Hour, ""); err != nil {
		m.answer(c, err)
		return
	}
	if err := password.Validate(req.Password, m.cfg.PasswordMin); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	if len(req.Token) != 43 {
		m.answer(c, errInvalid)
		return
	}
	var snapshot EmailToken
	err := m.db.WithContext(c.Request.Context()).Where("token_hash=? AND purpose='reset' AND used_at IS NULL AND expires_at>now()", digest(req.Token)).First(&snapshot).Error
	if err != nil || snapshot.UserID == nil {
		m.answer(c, errInvalid)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), m.cfg.HashCost)
	if err != nil {
		m.answer(c, err)
		return
	}
	err = m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		u, err := m.lockedUser(tx, *snapshot.UserID)
		if err != nil {
			return err
		}
		t, err := token(tx, req.Token, "reset")
		if err != nil {
			return err
		}
		if t.UserID == nil || *t.UserID != u.ID || t.PasswordHash != u.PasswordHash {
			return errInvalid
		}
		identity, err := m.identity.GetIdentity(c.Request.Context(), u.ID)
		if err != nil {
			return err
		}
		if !strings.EqualFold(identity.Email, t.Email) {
			return errInvalid
		}
		if err := tx.Model(u).Updates(map[string]any{"password_hash": string(hash), "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		v := Security{UserID: u.ID, VerifiedEmail: t.Email, VerifiedAt: time.Now()}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}}, DoUpdates: clause.AssignmentColumns([]string{"verified_email", "verified_at"})}).Create(&v).Error; err != nil {
			return err
		}
		return m.invalidate(tx, u.ID)
	})
	if err != nil {
		m.answer(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Пароль изменён. Все сессии завершены. Войдите с новым паролем."})
}

func (m *Manager) ChangePassword(c *gin.Context) {
	p, ok := m.current(c)
	if !ok {
		return
	}
	var req struct {
		Current  string `json:"current_password"`
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
	if err := m.admit(c, "change-password", 5, 100, time.Hour, strings.TrimSpace(c.GetHeader("Authorization"))); err != nil {
		m.answer(c, err)
		return
	}
	if err := password.Validate(req.Password, m.cfg.PasswordMin); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), m.cfg.HashCost)
	if err != nil {
		m.answer(c, err)
		return
	}
	err = m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		u, err := m.lockedUser(tx, p.UserID)
		if err != nil {
			return err
		}
		if _, err = m.ValidateAccess(c.Request.Context(), bearer(c)); err != nil {
			return err
		}
		if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Current)) != nil {
			return errCurrentPassword
		}
		if err := tx.Model(u).Updates(map[string]any{"password_hash": string(hash), "updated_at": time.Now()}).Error; err != nil {
			return err
		}
		return m.invalidate(tx, u.ID)
	})
	if err != nil {
		m.answer(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Пароль изменён. Войдите заново на всех устройствах."})
}

func (m *Manager) ChangeEmail(c *gin.Context) {
	p, ok := m.current(c)
	if !ok {
		return
	}
	var req struct {
		Current string `json:"current_password"`
		Email   string `json:"email"`
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
		m.answer(c, err)
		return
	}
	if err = m.admit(c, "change-email", 5, 100, time.Hour, email); err != nil {
		m.answer(c, err)
		return
	}
	err = m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		u, err := m.lockedUser(tx, p.UserID)
		if err != nil {
			return err
		}
		if _, err = m.ValidateAccess(c.Request.Context(), bearer(c)); err != nil {
			return err
		}
		if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Current)) != nil {
			return errCurrentPassword
		}
		existing, find := m.identity.FindIdentity(c.Request.Context(), email, ffid.LookupEmail)
		if find == nil && existing.ID != u.ID {
			return errInvalid
		}
		if find != nil && !errors.Is(find, ffid.ErrNotFound) {
			return find
		}
		// enqueue invalidates every previous change-email token for this user.
		return m.enqueue(tx, "email", nil, &u.ID, email, u.PasswordHash, c.ClientIP())
	})
	if err != nil {
		m.answer(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Подтвердите новый адрес по ссылке в письме. До этого действует прежний email."})
}
func (m *Manager) ConfirmEmail(c *gin.Context) {
	p, ok := m.current(c)
	if !ok {
		return
	}
	var req struct {
		Token string `json:"token"`
	}
	if !bind(c, &req) {
		return
	}
	if err := m.admit(c, "confirm-email", 10, 100, time.Hour, ""); err != nil {
		m.answer(c, err)
		return
	}
	err := m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		u, err := m.lockedUser(tx, p.UserID)
		if err != nil {
			return err
		}
		if _, err = m.ValidateAccess(c.Request.Context(), bearer(c)); err != nil {
			return err
		}
		t, err := token(tx, req.Token, "email")
		if err != nil {
			return err
		}
		if t.UserID == nil || *t.UserID != u.ID || t.PasswordHash != u.PasswordHash {
			return errInvalid
		}
		_, err = m.identity.UpdateIdentity(c.Request.Context(), u.ID, &t.Email, nil)
		if err != nil {
			return err
		}
		if err := tx.Model(&Security{}).Where("user_id=?", u.ID).Updates(map[string]any{"verified_email": t.Email, "verified_at": time.Now()}).Error; err != nil {
			return err
		}
		return m.invalidate(tx, u.ID)
	})
	if err != nil {
		m.answer(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Новый email подтверждён. Войдите заново."})
}

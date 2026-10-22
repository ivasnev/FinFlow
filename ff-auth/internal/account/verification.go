package account

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pendingRequest struct {
	PendingToken string `json:"pending_token"`
}

func (m *Manager) Resend(c *gin.Context) {
	var req pendingRequest
	if !bind(c, &req) {
		return
	}
	if err := m.admit(c, "resend", 10, 200, time.Hour, ""); err != nil {
		m.answer(c, err)
		return
	}
	err := m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		p, err := m.pending(tx, req.PendingToken)
		if err != nil {
			return err
		}
		return m.enqueue(tx, "verify", p, p.UserID, p.Email, p.PasswordHash, c.ClientIP())
	})
	if err != nil {
		m.answer(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Письмо поставлено в очередь. Проверьте почту."})
}
func (m *Manager) CorrectEmail(c *gin.Context) {
	var req struct {
		PendingToken string `json:"pending_token"`
		Email        string `json:"email"`
		Password     string `json:"password"`
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
	if err = m.admit(c, "correct-email", 5, 100, time.Hour, email); err != nil {
		m.answer(c, err)
		return
	}
	var p *Pending
	err = m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		p, err = m.pending(tx, req.PendingToken)
		if err != nil {
			return err
		}
		if bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(req.Password)) != nil {
			return errCurrentPassword
		}
		if p.UserID != nil && !p.ExistingUser {
			return errInvalid
		}
		_, find := m.identity.FindIdentity(c.Request.Context(), email, ffid.LookupEmail)
		if find == nil {
			return errInvalid
		}
		if !errors.Is(find, ffid.ErrNotFound) {
			return find
		}
		if err := tx.Model(p).Update("email", email).Error; err != nil {
			return err
		}
		p.Email = email
		return m.enqueue(tx, "verify", p, p.UserID, email, p.PasswordHash, c.ClientIP())
	})
	if err != nil {
		m.answer(c, err)
		return
	}
	pendingResponse(c, p, req.PendingToken)
}
func (m *Manager) PendingLogout(c *gin.Context) {
	var req pendingRequest
	if !bind(c, &req) {
		return
	}
	err := m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		p, err := m.pending(tx, req.PendingToken)
		if err != nil {
			return err
		}
		// An in-progress provision keeps its stable ID for recovery; revoke just the browser capability.
		return tx.Model(p).Update("secret_hash", digest(randomToken())).Error
	})
	if err != nil {
		m.answer(c, err)
		return
	}
	c.Status(204)
}
func token(tx *gorm.DB, raw, purpose string) (*EmailToken, error) {
	if len(raw) != 43 {
		return nil, errInvalid
	}
	var t EmailToken
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("token_hash=? AND purpose=? AND used_at IS NULL AND expires_at>now()", digest(raw), purpose).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errInvalid
	}
	return &t, err
}
func (m *Manager) verifyPair(tx *gorm.DB, secret, raw string) (*Pending, *EmailToken, error) {
	p, err := m.pending(tx, secret)
	if err != nil {
		return nil, nil, err
	}
	t, err := token(tx, raw, "verify")
	if err != nil {
		return nil, nil, err
	}
	if t.PendingID == nil || *t.PendingID != p.ID || t.Email != p.Email || t.PasswordHash != p.PasswordHash {
		return nil, nil, errInvalid
	}
	return p, t, nil
}

func (m *Manager) Verify(c *gin.Context) {
	var req struct {
		PendingToken string `json:"pending_token"`
		Token        string `json:"token"`
	}
	if !bind(c, &req) {
		return
	}
	if err := m.admit(c, "verify", 20, 500, time.Hour, ""); err != nil {
		m.answer(c, err)
		return
	}
	// Reserve a stable credentials ID before contacting ff-id. A timeout/rollback must not allocate another profile ID on retry.
	err := m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		p, _, err := m.verifyPair(tx, req.PendingToken, req.Token)
		if err != nil {
			return err
		}
		if p.UserID != nil {
			return nil
		}
		u := credentials{Email: p.Email, Nickname: p.Nickname, PasswordHash: p.PasswordHash, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := tx.Create(&u).Error; err != nil {
			return err
		}
		if err := tx.Exec(`INSERT INTO user_roles(user_id,role_id) SELECT ?,id FROM roles WHERE name='user'`, u.ID).Error; err != nil {
			return err
		}
		return tx.Model(p).Update("user_id", u.ID).Error
	})
	if err != nil {
		m.answer(c, err)
		return
	}
	var response *service.AccessDataParams
	err = m.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		var snapshot Pending
		if err := tx.Where("secret_hash=?", digest(req.PendingToken)).First(&snapshot).Error; err != nil {
			return errUnauthorized
		}
		if snapshot.UserID == nil {
			return errInvalid
		}
		u, err := m.lockedUser(tx, *snapshot.UserID)
		if err != nil {
			return err
		}
		p, t, err := m.verifyPair(tx, req.PendingToken, req.Token)
		if err != nil {
			return err
		}
		if u.PasswordHash != p.PasswordHash {
			return errInvalid
		}
		identity, err := m.identity.GetIdentity(c.Request.Context(), u.ID)
		if !p.ExistingUser && errors.Is(err, ffid.ErrNotFound) {
			identity, err = m.identity.RegisterUser(c.Request.Context(), &ffid.RegisterUserRequest{UserID: u.ID, Email: p.Email, Nickname: p.Nickname, Name: p.Name, Phone: p.Phone})
			if err != nil {
				// ff-id can commit before its response is lost. Read back only our reserved ID; never attach by email.
				identity, err = m.identity.GetIdentity(c.Request.Context(), u.ID)
			}
		}
		if err != nil {
			return err
		}
		if p.ExistingUser {
			if !strings.EqualFold(identity.Email, p.OriginalEmail) {
				return errInvalid
			}
			if !strings.EqualFold(identity.Email, p.Email) {
				identity, err = m.identity.UpdateIdentity(c.Request.Context(), u.ID, &p.Email, nil)
				if err != nil {
					return err
				}
			}
		}
		if !strings.EqualFold(identity.Email, p.Email) || identity.Nickname != p.Nickname {
			return errInvalid
		}
		v := Security{UserID: u.ID, VerifiedEmail: p.Email, VerifiedAt: time.Now()}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}}, DoUpdates: clause.AssignmentColumns([]string{"verified_email", "verified_at"})}).Create(&v).Error; err != nil {
			return err
		}
		if err := tx.Exec(`DELETE FROM sessions WHERE user_id=?`, u.ID).Error; err != nil {
			return err
		}
		access, err := m.session(tx, u, identity, c.ClientIP())
		if err != nil {
			return err
		}
		if err := m.recordLogin(tx, u.ID, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
			return err
		}
		response = access
		if err := tx.Model(t).Update("used_at", time.Now()).Error; err != nil {
			return err
		}
		// All pending capabilities for this account are invalid once ownership has been established.
		return tx.Where("user_id=?", u.ID).Delete(&Pending{}).Error
	})
	if err != nil {
		m.answer(c, err)
		return
	}
	// Stored as any only to keep the transaction result separate from capability records.
	authResponse(c, response)
}

func (m *Manager) Account(c *gin.Context) {
	p, ok := m.current(c)
	if !ok {
		return
	}
	identity, err := m.identity.GetIdentity(c.Request.Context(), p.UserID)
	if err != nil {
		m.answer(c, err)
		return
	}
	c.JSON(200, gin.H{"email": identity.Email, "email_verified": true})
}

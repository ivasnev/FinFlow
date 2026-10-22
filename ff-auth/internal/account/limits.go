package account

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

func (m *Manager) limit(tx *gorm.DB, key string, n int, window time.Duration) error {
	if n <= 0 {
		return errLimited
	}
	bucket := time.Now().UTC().UnixNano() / int64(window)
	key = fmt.Sprintf("%s:%d", key, bucket)
	return m.counter(tx, key, n, time.Now().Add(2*window))
}
func (m *Manager) counter(tx *gorm.DB, key string, n int, expiry time.Time) error {
	key = m.key(key)
	result := tx.Exec(`INSERT INTO account_limits(key,count,expires_at) VALUES(?,1,?) ON CONFLICT(key) DO UPDATE SET count=account_limits.count+1 WHERE account_limits.count < ?`, key, expiry, n)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errLimited
	}
	return nil
}

// Counters for attempts are committed even when a narrower quota rejects the operation.
func (m *Manager) admit(c *gin.Context, action string, perIP, global int, window time.Duration, identity string) error {
	tx := m.db.WithContext(c.Request.Context())
	if err := m.limit(tx, action+":global", global, window); err != nil {
		return err
	}
	if err := m.limit(tx, action+":ip:"+c.ClientIP(), perIP, window); err != nil {
		return err
	}
	if identity != "" {
		return m.limit(tx, action+":identity:"+identity, perIP, window)
	}
	return nil
}
func (m *Manager) monthLimit(tx *gorm.DB, key string, n int) error {
	now := time.Now().UTC()
	expiry := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC).Add(time.Hour)
	return m.counter(tx, key+":"+now.Format("2006-01"), n, expiry)
}
func (m *Manager) reserveMail(tx *gorm.DB, email, ip string) error {
	result := tx.Exec(`INSERT INTO account_limits(key,count,expires_at) VALUES(?,1,now()+interval '60 seconds') ON CONFLICT(key) DO UPDATE SET count=1,expires_at=EXCLUDED.expires_at WHERE account_limits.expires_at<=now()`, m.key("mail:cooldown:"+email))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errLimited
	}
	if err := m.monthLimit(tx, "mail:global:month", m.cfg.MailMonthly); err != nil {
		return err
	}
	for _, r := range []struct {
		key    string
		n      int
		window time.Duration
	}{
		{"mail:global:day", m.cfg.MailDaily, 24 * time.Hour},
		{"mail:recipient:hour:" + email, 3, time.Hour},
		{"mail:recipient:day:" + email, 10, 24 * time.Hour}, {"mail:ip:day:" + ip, 20, 24 * time.Hour},
	} {
		if err := m.limit(tx, r.key, r.n, r.window); err != nil {
			return err
		}
	}
	return nil
}

package account

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type delivery struct{ To, Subject, Body string }

func (m *Manager) enqueue(tx *gorm.DB, purpose string, p *Pending, userID *int64, email, passwordHash, ip string) error {
	if err := m.reserveMail(tx, email, ip); err != nil {
		return err
	}
	previous := tx.Where("purpose=?", purpose)
	if p != nil {
		previous = previous.Where("pending_id=?", p.ID)
	} else {
		previous = previous.Where("user_id=?", userID)
	}
	if err := previous.Delete(&EmailToken{}).Error; err != nil {
		return err
	}
	raw := randomToken()
	id := uuid.New()
	t := EmailToken{ID: id, TokenHash: digest(raw), Purpose: purpose, UserID: userID, Email: email, PasswordHash: passwordHash, ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now()}
	if p != nil {
		t.PendingID = &p.ID
	}
	path, subject := "/verify-email", "Подтвердите почту в FinFlow"
	if purpose == "reset" {
		path = "/reset-password"
		subject = "Восстановление пароля FinFlow"
	}
	if purpose == "email" {
		path = "/confirm-email"
		subject = "Подтвердите новый адрес FinFlow"
	}
	body := "Откройте ссылку и подтвердите действие:\n\n" + strings.TrimRight(m.cfg.AppURL, "/") + path + "#token=" + raw + "\n\nСсылка действует один час. Если вы не запрашивали это письмо, ничего не подтверждайте."
	if purpose == "verify" {
		body += "\nПодтверждение требует браузер, в котором вы регистрировались, либо вход с выбранным вами паролем. Не подтверждайте чужую регистрацию."
	}
	plain, err := json.Marshal(delivery{To: email, Subject: subject, Body: body})
	if err != nil {
		return err
	}
	aead := m.cfg.cipher()
	nonce := make([]byte, aead.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return err
	}
	encrypted := append(nonce, aead.Seal(nil, nonce, plain, []byte(id.String()))...)
	if err := tx.Create(&t).Error; err != nil {
		return err
	}
	return tx.Create(&Outbox{ID: id, Payload: encrypted, AvailableAt: time.Now(), CreatedAt: time.Now()}).Error
}

// Run runs one bounded outbox worker and bounded cleanup. SKIP LOCKED plus a fencing token permits multiple replicas.
func (m *Manager) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	cleanup := time.NewTicker(time.Minute)
	defer cleanup.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = m.DeliverOne(ctx)
		case <-cleanup.C:
			_ = m.Cleanup(ctx)
		}
	}
}
func (m *Manager) DeliverOne(ctx context.Context) error {
	var row Outbox
	lease := uuid.New()
	claimed := false
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("sent_at IS NULL AND failed_at IS NULL AND available_at<=now() AND (lease_until IS NULL OR lease_until<now())").Order("available_at").First(&row).Error
		if err != nil {
			return err
		}
		var n int64
		if err := tx.Model(&EmailToken{}).Where("id=? AND used_at IS NULL AND expires_at>now()", row.ID).Count(&n).Error; err != nil {
			return err
		}
		if n != 1 {
			return tx.Delete(&row).Error
		}
		if row.Attempts >= 5 {
			return tx.Model(&row).Updates(map[string]any{"failed_at": time.Now(), "payload": []byte{}}).Error
		}
		// Bound actual SMTP attempts too: retries may produce duplicate deliveries after a lost acknowledgement.
		if err := m.limit(tx, "smtp:day", m.cfg.MailDaily, 24*time.Hour); err != nil {
			return err
		}
		if err := m.monthLimit(tx, "smtp:month", m.cfg.MailMonthly); err != nil {
			return err
		}
		result := tx.Model(&row).Updates(map[string]any{"lease_token": lease, "lease_until": time.Now().Add(30 * time.Second), "attempts": gorm.Expr("attempts+1")})
		claimed = result.Error == nil
		return result.Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if !claimed || row.ID == uuid.Nil || len(row.Payload) == 0 || row.Attempts >= 5 {
		return nil
	}
	aead := m.cfg.cipher()
	var payload delivery
	if len(row.Payload) < aead.NonceSize() {
		return errors.New("invalid encrypted mail payload")
	}
	plain, err := aead.Open(nil, row.Payload[:aead.NonceSize()], row.Payload[aead.NonceSize():], []byte(row.ID.String()))
	if err == nil {
		err = json.Unmarshal(plain, &payload)
	}
	if err == nil {
		err = m.send(ctx, payload)
	}
	update := map[string]any{"lease_until": nil, "lease_token": nil}
	if err == nil {
		update["sent_at"] = time.Now()
		update["payload"] = []byte{}
	} else {
		update["available_at"] = time.Now().Add(time.Minute * time.Duration(1<<row.Attempts))
	}
	result := m.db.WithContext(ctx).Model(&Outbox{}).Where("id=? AND lease_token=?", row.ID, lease).Updates(update)
	if result.Error != nil {
		return result.Error
	}
	return err
}
func (m *Manager) send(ctx context.Context, p delivery) error {
	host, _, _ := net.SplitHostPort(m.cfg.SMTPAddr)
	deadlineCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(deadlineCtx, "tcp", m.cfg.SMTPAddr)
	if err != nil {
		return err
	}
	defer conn.Close()
	deadline, _ := deadlineCtx.Deadline()
	if err := conn.SetDeadline(deadline); err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if !m.cfg.Local {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return errors.New("SMTP requires STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if m.cfg.SMTPUser != "" {
		if err := client.Auth(smtp.PlainAuth("", m.cfg.SMTPUser, m.cfg.SMTPPassword, host)); err != nil {
			return err
		}
	}
	from, err := mail.ParseAddress(m.cfg.From)
	if err != nil {
		return err
	}
	to, err := mail.ParseAddress(p.To)
	if err != nil {
		return err
	}
	if err = client.Mail(from.Address); err != nil {
		return err
	}
	if err = client.Rcpt(to.Address); err != nil {
		return err
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	header := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n", m.cfg.From, to.Address, mime.QEncoding.Encode("utf-8", p.Subject))
	_, err = io.WriteString(writer, header+strings.ReplaceAll(p.Body, "\n", "\r\n"))
	if err != nil {
		return err
	}
	if err = writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func (m *Manager) Cleanup(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// No unbounded DELETE or vacuum in the request path. Provisioned profiles and existing user data are preserved.
	for _, q := range []string{
		`DELETE FROM account_pending WHERE id IN (SELECT id FROM account_pending WHERE expires_at<now() ORDER BY expires_at LIMIT 500)`,
		`DELETE FROM account_email_tokens WHERE id IN (SELECT id FROM account_email_tokens WHERE expires_at<now() ORDER BY expires_at LIMIT 500)`,
		`DELETE FROM account_captcha WHERE id IN (SELECT id FROM account_captcha WHERE expires_at<now() ORDER BY expires_at LIMIT 500)`,
		`DELETE FROM account_limits WHERE key IN (SELECT key FROM account_limits WHERE expires_at<now() ORDER BY expires_at LIMIT 500)`,
		`DELETE FROM account_access WHERE token_hash IN (SELECT token_hash FROM account_access WHERE expires_at<now() ORDER BY expires_at LIMIT 500)`,
	} {
		// Drain up to 10,000 rows per table/tick: faster than admitted creation,
		// while every statement and the whole cleanup remain bounded.
		for batch := 0; batch < 20; batch++ {
			result := m.db.WithContext(ctx).Exec(q)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected < 500 {
				break
			}
		}
	}
	return nil
}

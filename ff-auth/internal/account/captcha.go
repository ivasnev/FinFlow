package account

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type challenge struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	AnswerHash string
	ExpiresAt  time.Time
}

func (challenge) TableName() string { return "account_captcha" }
func (m *Manager) Captcha(c *gin.Context) {
	if m.cfg.CaptchaProvider != "image" {
		c.JSON(400, gin.H{"error": "Используйте SmartCaptcha."})
		return
	}
	if err := m.admit(c, "captcha", 20, 300, time.Minute, ""); err != nil {
		m.answer(c, err)
		return
	}
	id := uuid.New()
	digits := ""
	for range 6 {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			m.answer(c, err)
			return
		}
		digits += n.String()
	}
	record := challenge{ID: id, AnswerHash: m.key(id.String() + ":" + digits), ExpiresAt: time.Now().Add(3 * time.Minute)}
	if err := m.db.WithContext(c.Request.Context()).Create(&record).Error; err != nil {
		m.answer(c, err)
		return
	}
	// Deliberately local-only adapter. Production is fail-closed and requires the cloud verifier.
	picture := image.NewRGBA(image.Rect(0, 0, 240, 76))
	draw.Draw(picture, picture.Bounds(), image.NewUniform(color.RGBA{244, 248, 237, 255}), image.Point{}, draw.Src)
	segments := []byte{0x3f, 0x06, 0x5b, 0x4f, 0x66, 0x6d, 0x7d, 0x07, 0x7f, 0x6f}
	for i, digit := range digits {
		x := 12 + i*38
		y := 12
		rects := []image.Rectangle{image.Rect(x+5, y, x+23, y+4), image.Rect(x+23, y+4, x+27, y+23), image.Rect(x+23, y+27, x+27, y+46), image.Rect(x+5, y+46, x+23, y+50), image.Rect(x+1, y+27, x+5, y+46), image.Rect(x+1, y+4, x+5, y+23), image.Rect(x+5, y+23, x+23, y+27)}
		for j, r := range rects {
			if segments[digit-'0']&(1<<j) != 0 {
				draw.Draw(picture, r, image.NewUniform(color.RGBA{42, 65, 30, 255}), image.Point{}, draw.Src)
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, picture); err != nil {
		m.answer(c, err)
		return
	}
	c.JSON(200, gin.H{"id": id, "image": "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), "expires_at": record.ExpiresAt})
}
func (m *Manager) captcha(ctx context.Context, token, ip string) error {
	if len(token) == 0 || len(token) > 8192 {
		return errCaptcha
	}
	if m.cfg.CaptchaProvider == "image" {
		parts := strings.Split(token, ":")
		if len(parts) != 2 || len(parts[1]) != 6 {
			return errCaptcha
		}
		id, err := uuid.Parse(parts[0])
		if err != nil {
			return errCaptcha
		}
		// Consume on any attempt, including an incorrect answer. Expiry and single-use survive restarts.
		var value challenge
		result := m.db.WithContext(ctx).Raw(`DELETE FROM account_captcha WHERE id=? AND expires_at>now() RETURNING *`, id).Scan(&value)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 || value.AnswerHash != m.key(token) {
			return errCaptcha
		}
		return nil
	}
	if err := m.counter(m.db.WithContext(ctx), "captcha:used:"+digest(token), 1, time.Now().Add(10*time.Minute)); err != nil {
		return errCaptcha
	}
	form := url.Values{"secret": {m.cfg.CaptchaSecret}, "token": {token}, "ip": {ip}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://smartcaptcha.cloud.yandex.ru/validate", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := m.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var result struct {
		Status string `json:"status"`
		Host   string `json:"host"`
	}
	if resp.StatusCode != 200 || json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&result) != nil {
		return errCaptcha
	}
	origin, _ := url.Parse(m.cfg.AppURL)
	if result.Status != "ok" || result.Host != origin.Host {
		return errCaptcha
	}
	return nil
}

// Package account implements the public account lifecycle. Pending credentials never grant application access.
package account

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"net"
	"net/mail"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Local                                                  bool
	Secret, AppURL, SMTPAddr, SMTPUser, SMTPPassword, From string
	CaptchaProvider, CaptchaSiteKey, CaptchaSecret         string
	PasswordMin, HashCost                                  int
	AccessTTL, RefreshTTL                                  time.Duration
	MailDaily, MailMonthly                                 int
}

func ConfigFromEnv(minimum, cost, accessMinutes, refreshMinutes int) Config {
	c := Config{Local: os.Getenv("ACCOUNT_LOCAL_MODE") == "true", Secret: os.Getenv("ACCOUNT_SECRET"), AppURL: os.Getenv("ACCOUNT_APP_URL"), SMTPAddr: os.Getenv("ACCOUNT_SMTP_ADDR"), SMTPUser: os.Getenv("ACCOUNT_SMTP_USER"), SMTPPassword: os.Getenv("ACCOUNT_SMTP_PASSWORD"), From: os.Getenv("ACCOUNT_MAIL_FROM"), CaptchaProvider: os.Getenv("ACCOUNT_CAPTCHA_PROVIDER"), CaptchaSiteKey: os.Getenv("SMARTCAPTCHA_SITE_KEY"), CaptchaSecret: os.Getenv("SMARTCAPTCHA_SECRET_KEY"), PasswordMin: minimum, HashCost: cost, AccessTTL: time.Duration(accessMinutes) * time.Minute, RefreshTTL: time.Duration(refreshMinutes) * time.Minute, MailDaily: 100, MailMonthly: 2000}
	if c.Local {
		if c.AppURL == "" {
			c.AppURL = "http://127.0.0.1:3000"
		}
		if c.SMTPAddr == "" {
			c.SMTPAddr = "mailpit:1025"
		}
		if c.From == "" {
			c.From = "FinFlow <noreply@finflow.test>"
		}
		if c.CaptchaProvider == "" {
			c.CaptchaProvider = "image"
		}
	}
	if c.CaptchaProvider == "" {
		c.CaptchaProvider = "smartcaptcha"
	}
	if n, err := strconv.Atoi(os.Getenv("ACCOUNT_MAIL_DAILY_LIMIT")); err == nil {
		c.MailDaily = n
	}
	if n, err := strconv.Atoi(os.Getenv("ACCOUNT_MAIL_MONTHLY_LIMIT")); err == nil {
		c.MailMonthly = n
	}
	return c
}

func (c Config) validate() error {
	u, err := url.Parse(c.AppURL)
	if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
		return errors.New("ACCOUNT_APP_URL must be an application origin")
	}
	if u.Scheme != "https" && !(c.Local && u.Scheme == "http" && (u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost")) {
		return errors.New("ACCOUNT_APP_URL requires HTTPS (loopback HTTP is allowed in local mode)")
	}
	if len(c.Secret) < 32 {
		return errors.New("ACCOUNT_SECRET must contain at least 32 characters and remain stable across restarts")
	}
	if strings.ContainsAny(c.From, "\r\n") || c.From == "" {
		return errors.New("ACCOUNT_MAIL_FROM is required")
	}
	if _, err := mail.ParseAddress(c.From); err != nil {
		return errors.New("ACCOUNT_MAIL_FROM must be a valid sender")
	}
	host, _, err := net.SplitHostPort(c.SMTPAddr)
	if err != nil {
		return errors.New("ACCOUNT_SMTP_ADDR must be host:port")
	}
	if c.Local && host != "mailpit" && host != "localhost" && net.ParseIP(host) == nil {
		return errors.New("local SMTP must use mailpit or a loopback address")
	}
	if c.Local && net.ParseIP(host) != nil && !net.ParseIP(host).IsLoopback() {
		return errors.New("local SMTP must be loopback")
	}
	if c.CaptchaProvider != "smartcaptcha" && !(c.Local && c.CaptchaProvider == "image") {
		return errors.New("production requires smartcaptcha; local image captcha is available only in local mode")
	}
	if c.CaptchaProvider == "smartcaptcha" && (c.CaptchaSiteKey == "" || c.CaptchaSecret == "") {
		return errors.New("SmartCaptcha keys are required")
	}
	if c.PasswordMin < 8 || c.HashCost < 4 || c.HashCost > 16 {
		return errors.New("password minimum must be >=8 and bcrypt cost between 4 and 16")
	}
	if c.MailDaily < 1 || c.MailMonthly < 1 || c.AccessTTL <= 0 || c.RefreshTTL <= 0 {
		return errors.New("positive mail limits and session durations are required")
	}
	return nil
}

func (c Config) cipher() cipher.AEAD {
	key := sha256.Sum256([]byte(c.Secret))
	block, _ := aes.NewCipher(key[:])
	aead, _ := cipher.NewGCM(block)
	return aead
}

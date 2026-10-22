package handler

import "github.com/gin-gonic/gin"

func (h *ServerHandler) GetAccountConfig(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.ConfigHTTP(c)
}
func (h *ServerHandler) GetCaptcha(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.Captcha(c)
}
func (h *ServerHandler) ResendVerification(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.Resend(c)
}
func (h *ServerHandler) CorrectPendingEmail(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.CorrectEmail(c)
}
func (h *ServerHandler) VerifyEmail(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.Verify(c)
}
func (h *ServerHandler) LogoutPending(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.PendingLogout(c)
}
func (h *ServerHandler) ForgotPassword(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.Forgot(c)
}
func (h *ServerHandler) ResetPassword(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.Reset(c)
}
func (h *ServerHandler) ChangePassword(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.ChangePassword(c)
}
func (h *ServerHandler) RequestEmailChange(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.ChangeEmail(c)
}
func (h *ServerHandler) ConfirmEmailChange(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.ConfirmEmail(c)
}
func (h *ServerHandler) Introspect(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.Introspect(c)
}
func (h *ServerHandler) GetAccountSecurity(c *gin.Context) {
	if h.Account == nil {
		c.JSON(503, gin.H{"error": "account security unavailable"})
		return
	}
	h.Account.Account(c)
}

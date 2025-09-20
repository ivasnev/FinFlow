package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/interfaces"
)

// SessionHandler обработчик запросов сессий
type SessionHandler struct {
	sessionService      interfaces.SessionService
	loginHistoryService interfaces.LoginHistoryService
}

// NewSessionHandler создает новый обработчик сессий
func NewSessionHandler(
	sessionService interfaces.SessionService,
	loginHistoryService interfaces.LoginHistoryService,
) *SessionHandler {
	return &SessionHandler{
		sessionService:      sessionService,
		loginHistoryService: loginHistoryService,
	}
}

// GetUserSessions получает все сессии пользователя
func (h *SessionHandler) GetUserSessions(c *gin.Context) {
	// Заглушка для получения сессий
	c.JSON(http.StatusOK, gin.H{"sessions": []string{}})
}

// TerminateSession завершает сессию пользователя
func (h *SessionHandler) TerminateSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	// Заглушка для завершения сессии
	c.JSON(http.StatusOK, gin.H{"status": "session terminated", "session_id": sessionID})
}

// GetLoginHistory получает историю входов пользователя
func (h *SessionHandler) GetLoginHistory(c *gin.Context) {
	page := 1
	pageSize := 10

	// Заглушка для получения истории входов
	c.JSON(http.StatusOK, gin.H{
		"history": []string{},
		"page":    page,
		"size":    pageSize,
	})
}

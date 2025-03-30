package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-id/interfaces"
)

// SessionHandler обрабатывает запросы, связанные с сессиями пользователей
type SessionHandler struct {
	sessionService      interfaces.SessionService
	loginHistoryService interfaces.LoginHistoryService
}

// NewSessionHandler создает новый SessionHandler
func NewSessionHandler(
	sessionService interfaces.SessionService,
	loginHistoryService interfaces.LoginHistoryService,
) *SessionHandler {
	return &SessionHandler{
		sessionService:      sessionService,
		loginHistoryService: loginHistoryService,
	}
}

// GetUserSessions обрабатывает запрос на получение активных сессий пользователя
// @Summary Get user sessions
// @Description Get all active sessions for the current user
// @Tags sessions
// @Produce json
// @Success 200 {array} dto.SessionDTO
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /sessions [get]
func (h *SessionHandler) GetUserSessions(c *gin.Context) {
	// Получаем ID пользователя из контекста, установленного middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Преобразуем ID в int64
	var userIDInt64 int64
	switch v := userID.(type) {
	case int64:
		userIDInt64 = v
	case float64:
		userIDInt64 = int64(v)
	case string:
		var err error
		userIDInt64, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID format"})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	sessions, err := h.sessionService.GetUserSessions(c.Request.Context(), userIDInt64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// TerminateSession обрабатывает запрос на завершение сессии
// @Summary Terminate session
// @Description Terminate a specific session for the current user
// @Tags sessions
// @Produce json
// @Param id path string true "Session ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /sessions/{id} [delete]
func (h *SessionHandler) TerminateSession(c *gin.Context) {
	sessionIDStr := c.Param("id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session ID format"})
		return
	}

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Преобразуем ID в int64
	var userIDInt64 int64
	switch v := userID.(type) {
	case int64:
		userIDInt64 = v
	case float64:
		userIDInt64 = int64(v)
	case string:
		var err error
		userIDInt64, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID format"})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	err = h.sessionService.TerminateSession(c.Request.Context(), sessionID, userIDInt64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session terminated successfully"})
}

// GetLoginHistory обрабатывает запрос на получение истории входов пользователя
// @Summary Get login history
// @Description Get login history for the current user
// @Tags login-history
// @Produce json
// @Param limit query int false "Limit items per page" default(10)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} dto.LoginHistoryDTO
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /login-history [get]
func (h *SessionHandler) GetLoginHistory(c *gin.Context) {
	// Получаем параметры пагинации
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Преобразуем ID в int64
	var userIDInt64 int64
	switch v := userID.(type) {
	case int64:
		userIDInt64 = v
	case float64:
		userIDInt64 = int64(v)
	case string:
		var err error
		userIDInt64, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID format"})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	history, err := h.loginHistoryService.GetUserLoginHistory(c.Request.Context(), userIDInt64, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

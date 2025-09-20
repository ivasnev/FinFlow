package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/interfaces"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// UserHandler обработчик запросов пользователей
type UserHandler struct {
	userService interfaces.UserService
}

// NewUserHandler создает новый обработчик пользователей
func NewUserHandler(userService interfaces.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUserByNickname получает информацию о пользователе по никнейму
func (h *UserHandler) GetUserByNickname(c *gin.Context) {
	nickname := c.Param("nickname")

	// Заглушка для получения пользователя
	c.JSON(http.StatusOK, gin.H{"nickname": nickname})
}

// UpdateUser обновляет данные пользователя
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var userData models.UserUpdate
	if err := c.ShouldBindJSON(&userData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Заглушка для обновления пользователя
	c.JSON(http.StatusOK, gin.H{"status": "updated"})
}

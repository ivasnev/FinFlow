package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/interfaces"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// AuthHandler обработчик запросов аутентификации
type AuthHandler struct {
	authService interfaces.AuthService
}

// NewAuthHandler создает новый обработчик аутентификации
func NewAuthHandler(authService interfaces.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register обрабатывает запрос на регистрацию
func (h *AuthHandler) Register(c *gin.Context) {
	var user models.UserRegistration
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Заглушка для регистрации
	c.JSON(http.StatusCreated, gin.H{"status": "registered"})
}

// Login обрабатывает запрос на аутентификацию
func (h *AuthHandler) Login(c *gin.Context) {
	var credentials models.UserCredentials
	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Заглушка для входа
	c.JSON(http.StatusOK, gin.H{"status": "logged in"})
}

// RefreshToken обрабатывает запрос на обновление токенов
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Заглушка для обновления токенов
	c.JSON(http.StatusOK, gin.H{"status": "tokens refreshed"})
}

// Logout обрабатывает запрос на выход из системы
func (h *AuthHandler) Logout(c *gin.Context) {
	var request struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Заглушка для выхода
	c.JSON(http.StatusOK, gin.H{"status": "logged out"})
}

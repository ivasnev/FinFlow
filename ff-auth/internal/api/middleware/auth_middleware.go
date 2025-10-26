package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
)

// AuthMiddleware создает middleware для аутентификации запросов
func AuthMiddleware(authService service.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		// Проверяем формат токена
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Валидируем токен
		userID, roles, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Проверяем, имеет ли пользователь роль администратора
		isAdmin := false
		for _, role := range roles {
			if role == "admin" {
				isAdmin = true
				break
			}
		}

		// Создаем структуру с данными пользователя
		userData := auth.UserData{
			UserID:  userID,
			Roles:   roles,
			IsAdmin: isAdmin,
		}

		// Устанавливаем данные пользователя в контекст
		c.Set(string(auth.UserContextKey()), userData)

		c.Next()
	}
}

// RoleMiddleware создает middleware для проверки роли пользователя
func RoleMiddleware(role string, authService service.Auth) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем данные пользователя из контекста
		userData, exists := auth.GetUserData(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		// Проверяем наличие требуемой роли
		if !auth.HasRole(userData, role) {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

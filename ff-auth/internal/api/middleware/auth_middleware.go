package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
)

// AuthMiddleware создает middleware для аутентификации запросов
func AuthMiddleware(authService service.AuthServiceInterface) gin.HandlerFunc {
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

		// Устанавливаем данные пользователя в контекст
		c.Set("user_id", userID)
		c.Set("user_roles", roles)
		// Проверяем, имеет ли пользователь роль администратора
		isAdmin := false
		for _, role := range roles {
			if role == "admin" {
				isAdmin = true
				break
			}
		}
		c.Set("is_admin", isAdmin)

		c.Next()
	}
}

// RoleMiddleware создает middleware для проверки роли пользователя
func RoleMiddleware(role string, authService service.AuthServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем ID пользователя из контекста
		_, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		// Получаем роли пользователя из контекста
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user roles not found"})
			c.Abort()
			return
		}

		// Проверяем наличие требуемой роли
		hasRole := false
		for _, userRole := range userRoles.([]string) {
			if userRole == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

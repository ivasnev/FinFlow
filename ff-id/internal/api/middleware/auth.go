package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/interfaces"
	"github.com/ivasnev/FinFlow/ff-id/internal/models"
)

// AuthMiddleware создает middleware для проверки JWT-токена
func AuthMiddleware(authService interfaces.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		// Извлекаем токен из заголовка
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Проверяем токен
		userID, roles, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "details": err.Error()})
			c.Abort()
			return
		}

		// Устанавливаем данные пользователя в контекст
		c.Set("user_id", userID)
		c.Set("roles", roles)
		c.Next()
	}
}

// RoleMiddleware создает middleware для проверки ролей пользователя
func RoleMiddleware(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRoles, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		userRolesSlice, ok := userRoles.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid roles format"})
			c.Abort()
			return
		}

		// Проверяем, есть ли у пользователя хотя бы одна из требуемых ролей
		hasRole := false
		for _, requiredRole := range roles {
			for _, userRole := range userRolesSlice {
				if string(requiredRole) == userRole {
					hasRole = true
					break
				}
			}
			if hasRole {
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

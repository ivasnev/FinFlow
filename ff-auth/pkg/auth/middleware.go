package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type ValidateClient interface {
	ValidateToken(tokenStr string) (*TokenPayload, error)
}

// TokenPayload представляет содержимое токена
type TokenPayload struct {
	UserID int64    `json:"user_id"`
	Roles  []string `json:"roles"`
	Exp    int64    `json:"exp"`
}

// Token представляет структуру токена
type Token struct {
	Payload []byte `json:"payload"`
	Sig     []byte `json:"sig"`
}

// AuthMiddleware создает middleware для аутентификации запросов
func AuthMiddleware(client ValidateClient) gin.HandlerFunc {
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

		tokenStr := parts[1]

		// Валидируем токен
		payload, err := client.ValidateToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Устанавливаем данные пользователя в контекст
		c.Set("user_id", payload.UserID)
		c.Set("user_roles", payload.Roles)

		// Проверяем, имеет ли пользователь роль администратора
		isAdmin := false
		for _, role := range payload.Roles {
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
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем роли пользователя из контекста
		userRoles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Проверяем наличие требуемой роли
		hasRequiredRole := false
		for _, requiredRole := range roles {
			for _, userRole := range userRoles.([]string) {
				if userRole == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}

		c.Next()
	}
}

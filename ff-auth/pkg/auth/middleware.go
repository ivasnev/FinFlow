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

// contextKey - тип для ключей контекста
type contextKey string

// Ключ для хранения данных пользователя в контексте
const userContextKey contextKey = "user"

// UserData содержит данные пользователя из токена
type UserData struct {
	UserID  int64
	Roles   []string
	IsAdmin bool
}

// UserContextKey возвращает ключ для хранения данных пользователя в контексте
func UserContextKey() contextKey {
	return userContextKey
}

// GetUserData извлекает данные пользователя из контекста
func GetUserData(c *gin.Context) (*UserData, bool) {
	data, exists := c.Get(string(userContextKey))
	if !exists {
		return nil, false
	}
	userData, ok := data.(UserData)
	if !ok {
		return nil, false
	}
	return &userData, true
}

// HasRole проверяет, имеет ли пользователь указанную роль
func HasRole(userData *UserData, role string) bool {
	for _, r := range userData.Roles {
		if r == role {
			return true
		}
	}
	return false
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token", "error_description": err.Error()})
			c.Abort()
			return
		}

		// Проверяем, имеет ли пользователь роль администратора
		isAdmin := false
		for _, role := range payload.Roles {
			if role == "admin" {
				isAdmin = true
				break
			}
		}

		// Создаем структуру с данными пользователя
		userData := UserData{
			UserID:  payload.UserID,
			Roles:   payload.Roles,
			IsAdmin: isAdmin,
		}

		// Устанавливаем данные пользователя в контекст
		c.Set(string(userContextKey), userData)

		c.Next()
	}
}

// RoleMiddleware создает middleware для проверки роли пользователя
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем данные пользователя из контекста
		userData, exists := GetUserData(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Проверяем наличие требуемой роли
		hasRequiredRole := false
		for _, requiredRole := range roles {
			if HasRole(userData, requiredRole) {
				hasRequiredRole = true
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

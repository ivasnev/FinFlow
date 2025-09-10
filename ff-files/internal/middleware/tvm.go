package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type TVMClient interface {
	ValidateTicket(ticket string) (string, error)
}

func TVMAuth(tvmClient TVMClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
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

		// Проверяем тикет
		serviceID, err := tvmClient.ValidateTicket(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid ticket"})
			c.Abort()
			return
		}

		// Сохраняем ID сервиса в контексте
		c.Set("service_id", serviceID)
		c.Next()
	}
} 
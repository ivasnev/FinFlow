package middleware

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware добавляет заголовки CORS для поддержки кросс-доменных запросов
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Разрешаем запросы с фронтенда (можно настроить конкретные домены)
		c.Header("Access-Control-Allow-Origin", "*") // В продакшене заменить на конкретные домены
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")

		// Обработка preflight запросов (OPTIONS)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

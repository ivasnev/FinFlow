package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/service"
)

const (
	HeaderServiceID  = "X-Id-Service-Ticket"
	HeaderTicket     = "X-FF-Service-Ticket"
	ServiceIDContext = "service_id"
)

type TVMClient interface {
	GetPublicKey(serviceID int64) (string, error)
}

type TVMMiddleware struct {
	client TVMClient
}

func NewTVMMiddleware(client TVMClient) *TVMMiddleware {
	return &TVMMiddleware{
		client: client,
	}
}

func (m *TVMMiddleware) ValidateTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем ID сервиса из заголовка
		serviceID := c.GetHeader(HeaderServiceID)
		if serviceID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "service ID is required"})
			c.Abort()
			return
		}

		// TODO: Convert string ID to int64
		fromID := int64(0) // Placeholder

		// Получаем тикет из заголовка
		ticketStr := c.GetHeader(HeaderTicket)
		if ticketStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ticket is required"})
			c.Abort()
			return
		}

		// Декодируем тикет из base64
		ticketData, err := base64.StdEncoding.DecodeString(ticketStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid ticket format"})
			c.Abort()
			return
		}

		// Парсим тикет
		var ticket service.Ticket
		if err := json.Unmarshal(ticketData, &ticket); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid ticket data"})
			c.Abort()
			return
		}

		// Получаем публичный ключ сервиса
		publicKey, err := m.client.GetPublicKey(fromID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get service public key"})
			c.Abort()
			return
		}

		// Проверяем подпись тикета
		if err := service.ValidateTicketSignature(&ticket, publicKey); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid ticket signature"})
			c.Abort()
			return
		}

		// Сохраняем ID сервиса в контексте
		c.Set(ServiceIDContext, fromID)
		c.Next()
	}
}

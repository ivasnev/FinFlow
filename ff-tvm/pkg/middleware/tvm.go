package middleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/service"
)

const (
	HeaderServiceTicket = "X-FF-Service-Ticket"
	ServiceIDContext    = "service_id"
)

type TVMClient interface {
	GetPublicKey(serviceID int) (string, error)
}

// TVMMiddleware проверяет подпись и назначение билета принимающего сервиса.
type TVMMiddleware struct {
	client      TVMClient
	recipientID int
}

// NewTVMMiddleware создаёт проверку для явно указанного получателя.
func NewTVMMiddleware(client TVMClient, recipientID int) *TVMMiddleware {
	return &TVMMiddleware{
		client:      client,
		recipientID: recipientID,
	}
}

// ValidateTicket допускает только действующий билет для этого сервиса.
func (m *TVMMiddleware) ValidateTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем тикет из заголовка
		ticketStr := c.GetHeader(HeaderServiceTicket)
		if ticketStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "ticket is required"})
			c.Abort()
			return
		}

		// Парсим строку тикета
		parts := strings.Split(ticketStr, ":")
		if len(parts) != 3 || parts[0] != "serv" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid ticket format"})
			c.Abort()
			return
		}

		// Декодируем ID сервиса из base64
		serviceIDBytes, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid service ID format"})
			c.Abort()
			return
		}

		fromID, err := strconv.Atoi(string(serviceIDBytes))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "service ID must be int"})
			c.Abort()
			return
		}

		// Декодируем тикет из base64
		ticketData, err := base64.StdEncoding.DecodeString(parts[2])
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

		if m.recipientID <= 0 || fromID <= 0 || ticket.Payload.From != fromID ||
			ticket.Payload.To != m.recipientID || ticket.Payload.TTL <= time.Now().Unix() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid ticket claims"})
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

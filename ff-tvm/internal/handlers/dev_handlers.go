package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/service"
)

type DevHandlers struct {
	ticketService service.TicketService
}

func NewDevHandlers(ticketService service.TicketService) *DevHandlers {
	return &DevHandlers{
		ticketService: ticketService,
	}
}

type devTicketRequest struct {
	From int `json:"from" binding:"required"`
	To   int `json:"to" binding:"required"`
}

func (h *DevHandlers) GenerateDevTicket(c *gin.Context) {
	var req devTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Используем специальный секрет для разработки
	devSecret := "dev_secret_key" // В реальном приложении это должно быть в конфиге

	ticket, err := h.ticketService.GenerateTicket(c.Request.Context(), req.From, req.To, devSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *DevHandlers) RegisterRoutes(r *gin.Engine) {
	r.POST("/dev/ticket", h.GenerateDevTicket)
}

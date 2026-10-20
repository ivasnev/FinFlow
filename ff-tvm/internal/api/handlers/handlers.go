package handlers

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/service"
)

type Handlers struct {
	ticketService service.TicketService
	repo          service.ServiceRepository
	keyManager    service.KeyManager
}

func NewHandlers(ticketService service.TicketService, repo service.ServiceRepository, keyManager service.KeyManager) *Handlers {
	return &Handlers{
		ticketService: ticketService,
		repo:          repo,
		keyManager:    keyManager,
	}
}

type createTicketRequest struct {
	From   int    `json:"from" binding:"required"`
	To     int    `json:"to" binding:"required"`
	Secret string `json:"secret" binding:"required"`
}

func (h *Handlers) CreateTicket(c *gin.Context) {
	var req createTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.ticketService.GenerateTicket(c.Request.Context(), req.From, req.To, req.Secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Сериализуем тикет в JSON
	ticketJSON, err := json.Marshal(ticket)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to serialize ticket"})
		return
	}

	// Кодируем ID сервиса и тикет в base64
	serviceIDBase64 := base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(req.From)))
	ticketBase64 := base64.StdEncoding.EncodeToString(ticketJSON)

	// Формируем строку в формате "serv:base64(serviceID):base64(ticket)"
	ticketStr := fmt.Sprintf("serv:%s:%s", serviceIDBase64, ticketBase64)

	c.JSON(http.StatusOK, gin.H{"ticket": ticketStr})
}

func (h *Handlers) GetServicePublicKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service ID is required"})
		return
	}

	serviceID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid service ID format"})
		return
	}

	publicKey, err := h.repo.GetPublicKey(c.Request.Context(), serviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"public_key": publicKey})
}

type createServiceRequest struct {
	Name string `json:"name" binding:"required"`
}

type createServiceResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func (h *Handlers) CreateService(c *gin.Context) {
	var req createServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Генерируем пару ключей
	publicKey, privateKey, err := h.keyManager.GenerateKeyPair()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate key pair"})
		return
	}

	// Кодируем ключи в base64
	publicKeyStr := service.EncodeKey(publicKey)
	privateKeyStr := service.EncodeKey(privateKey)

	// Хешируем приватный ключ
	privateKeyHash := service.HashKey(privateKey)

	svc := &service.Service{
		Name:           req.Name,
		PublicKey:      publicKeyStr,
		PrivateKeyHash: privateKeyHash,
	}

	if err := h.repo.Create(c.Request.Context(), svc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createServiceResponse{
		ID:         svc.ID,
		Name:       svc.Name,
		PublicKey:  publicKeyStr,
		PrivateKey: privateKeyStr,
	})
}

type grantAccessRequest struct {
	From int `json:"from" binding:"required"`
	To   int `json:"to" binding:"required"`
}

func (h *Handlers) GrantAccess(c *gin.Context) {
	var req grantAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.GrantAccess(c.Request.Context(), req.From, req.To); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "access granted"})
}

type revokeAccessRequest struct {
	From int `json:"from" binding:"required"`
	To   int `json:"to" binding:"required"`
}

func (h *Handlers) RevokeAccess(c *gin.Context) {
	var req revokeAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.RevokeAccess(c.Request.Context(), req.From, req.To); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "access revoked"})
}

// RegisterRoutes регистрирует API и защищает изменения сервисов отдельным токеном администратора.
func (h *Handlers) RegisterRoutes(r *gin.Engine, adminToken string) {
	r.POST("/ticket", h.CreateTicket)
	r.GET("/service/:id/key", h.GetServicePublicKey)
	admin := r.Group("", requireAdmin(adminToken))
	admin.POST("/service", h.CreateService)
	admin.POST("/access/grant", h.GrantAccess)
	admin.POST("/access/revoke", h.RevokeAccess)
}

// requireAdmin закрывает доступ при отсутствии настройки или неверном токене.
func requireAdmin(token string) gin.HandlerFunc {
	expected := sha256.Sum256([]byte("Bearer " + token))
	return func(c *gin.Context) {
		provided := sha256.Sum256([]byte(c.GetHeader("Authorization")))
		if token == "" || subtle.ConstantTimeCompare(expected[:], provided[:]) != 1 {
			c.Header("WWW-Authenticate", "Bearer")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "admin authentication required"})
			return
		}
		c.Next()
	}
}

package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-tvm/internal/service"
	"net/http"
)

type TVMHandler struct {
	tvmService service.TVMService
}

func NewTVMHandler(tvmService service.TVMService) *TVMHandler {
	return &TVMHandler{
		tvmService: tvmService,
	}
}

type RegisterServiceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type GrantAccessRequest struct {
	SourceServiceID uint `json:"source_service_id" binding:"required"`
	TargetServiceID uint `json:"target_service_id" binding:"required"`
}

type IssueTicketRequest struct {
	SourceServiceID uint `json:"source_service_id" binding:"required"`
	TargetServiceID uint `json:"target_service_id" binding:"required"`
}

type ValidateTicketRequest struct {
	Ticket string `json:"ticket" binding:"required"`
}

func (h *TVMHandler) RegisterService(c *gin.Context) {
	var req RegisterServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service, err := h.tvmService.RegisterService(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		if err == service.ErrServiceExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register service"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          service.ID,
		"name":        service.Name,
		"description": service.Description,
		"public_key":  service.PublicKey,
	})
}

func (h *TVMHandler) GrantAccess(c *gin.Context) {
	var req GrantAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.tvmService.GrantAccess(c.Request.Context(), req.SourceServiceID, req.TargetServiceID)
	if err != nil {
		if err == service.ErrServiceNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to grant access"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "access granted successfully"})
}

func (h *TVMHandler) RevokeAccess(c *gin.Context) {
	var req GrantAccessRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.tvmService.RevokeAccess(c.Request.Context(), req.SourceServiceID, req.TargetServiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to revoke access"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "access revoked successfully"})
}

func (h *TVMHandler) IssueTicket(c *gin.Context) {
	var req IssueTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket, err := h.tvmService.IssueTicket(c.Request.Context(), req.SourceServiceID, req.TargetServiceID)
	if err != nil {
		switch err {
		case service.ErrServiceNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case service.ErrAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue ticket"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"ticket": ticket})
}

func (h *TVMHandler) ValidateTicket(c *gin.Context) {
	var req ValidateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, err := h.tvmService.ValidateTicket(c.Request.Context(), req.Ticket)
	if err != nil {
		switch err {
		case service.ErrInvalidTicket:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case service.ErrTicketExpired:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case service.ErrAccessDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to validate ticket"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"claims": claims})
}

func (h *TVMHandler) GetPublicKey(c *gin.Context) {
	serviceID := uint(c.GetInt("service_id"))
	publicKey, err := h.tvmService.GetPublicKey(c.Request.Context(), serviceID)
	if err != nil {
		if err == service.ErrServiceNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get public key"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"public_key": publicKey})
}

func (h *TVMHandler) RotateKeys(c *gin.Context) {
	serviceID := uint(c.GetInt("service_id"))
	err := h.tvmService.RotateKeys(c.Request.Context(), serviceID)
	if err != nil {
		if err == service.ErrServiceNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rotate keys"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "keys rotated successfully"})
} 
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// IconHandler обработчик для работы с иконками
type IconHandler struct {
	service service.IconServiceInterface
}

// NewIconHandler создает новый обработчик для работы с иконками
func NewIconHandler(service service.IconServiceInterface) *IconHandler {
	return &IconHandler{service: service}
}

// GetIcons возвращает список всех иконок
func (h *IconHandler) GetIcons(c *gin.Context) {
	icons, err := h.service.GetIcons(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.IconListResponse(icons))
}

// GetIconByID возвращает иконку по ID
func (h *IconHandler) GetIconByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID"})
		return
	}

	icon, err := h.service.GetIconByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.IconResponse(*icon))
}

// CreateIcon создает новую иконку
func (h *IconHandler) CreateIcon(c *gin.Context) {
	var iconDTO dto.IconFullDTO
	if err := c.ShouldBindJSON(&iconDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdIcon, err := h.service.CreateIcon(c.Request.Context(), &iconDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.IconResponse(*createdIcon))
}

// UpdateIcon обновляет существующую иконку
func (h *IconHandler) UpdateIcon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID"})
		return
	}

	var iconDTO dto.IconFullDTO
	if err := c.ShouldBindJSON(&iconDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedIcon, err := h.service.UpdateIcon(c.Request.Context(), uint(id), &iconDTO)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.IconResponse(*updatedIcon))
}

// DeleteIcon удаляет иконку по ID
func (h *IconHandler) DeleteIcon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID"})
		return
	}

	err = h.service.DeleteIcon(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

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
// @Summary Получить все иконки
// @Description Возвращает список всех доступных иконок в системе
// @Tags иконки
// @Accept json
// @Produce json
// @Success 200 {array} dto.IconResponse "Список иконок"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/manage/icons [get]
func (h *IconHandler) GetIcons(c *gin.Context) {
	icons, err := h.service.GetIcons(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.IconListResponse(icons))
}

// GetIconByID возвращает иконку по ID
// @Summary Получить иконку по ID
// @Description Возвращает информацию о конкретной иконке по её ID
// @Tags иконки
// @Accept json
// @Produce json
// @Param id path int true "ID иконки"
// @Success 200 {object} dto.IconResponse "Информация об иконке"
// @Failure 400 {object} map[string]string "Неверный формат ID"
// @Failure 404 {object} map[string]string "Иконка не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/manage/icons/{id} [get]
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
// @Summary Создать новую иконку
// @Description Создает новую иконку в системе
// @Tags иконки
// @Accept json
// @Produce json
// @Param icon body dto.IconFullDTO true "Данные иконки"
// @Success 201 {object} dto.IconResponse "Созданная иконка"
// @Failure 400 {object} map[string]string "Неверный формат данных запроса"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/manage/icons [post]
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
// @Summary Обновить иконку
// @Description Обновляет существующую иконку по ID
// @Tags иконки
// @Accept json
// @Produce json
// @Param id path int true "ID иконки"
// @Param icon body dto.IconFullDTO true "Данные иконки"
// @Success 200 {object} dto.IconResponse "Обновленная иконка"
// @Failure 400 {object} map[string]string "Неверный формат данных запроса"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/manage/icons/{id} [put]
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
// @Summary Удалить иконку
// @Description Удаляет иконку по ID
// @Tags иконки
// @Accept json
// @Produce json
// @Param id path int true "ID иконки"
// @Success 204 "Иконка успешно удалена"
// @Failure 400 {object} map[string]string "Неверный формат ID"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/manage/icons/{id} [delete]
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

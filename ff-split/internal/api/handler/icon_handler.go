package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
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
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/manage/icons [get]
func (h *IconHandler) GetIcons(c *gin.Context) {
	icons, err := h.service.GetIcons(c.Request.Context())
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении иконок: %w", err))
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
// @Failure 400 {object} errors.ErrorResponse "Неверный формат ID"
// @Failure 404 {object} errors.ErrorResponse "Иконка не найдена"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/manage/icons/{id} [get]
func (h *IconHandler) GetIconByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id", "неверный формат ID"))
		return
	}

	icon, err := h.service.GetIconByID(c.Request.Context(), uint(id))
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении иконки: %w", err))
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
// @Failure 400 {object} errors.ErrorResponse "Неверный формат данных запроса"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/manage/icons [post]
func (h *IconHandler) CreateIcon(c *gin.Context) {
	var iconDTO dto.IconFullDTO
	if err := c.ShouldBindJSON(&iconDTO); err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("request_body", err.Error()))
		return
	}

	createdIcon, err := h.service.CreateIcon(c.Request.Context(), &iconDTO)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при создании иконки: %w", err))
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
// @Failure 400 {object} errors.ErrorResponse "Неверный формат данных запроса"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/manage/icons/{id} [put]
func (h *IconHandler) UpdateIcon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id", "неверный формат ID"))
		return
	}

	var iconDTO dto.IconFullDTO
	if err := c.ShouldBindJSON(&iconDTO); err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("request_body", err.Error()))
		return
	}

	updatedIcon, err := h.service.UpdateIcon(c.Request.Context(), uint(id), &iconDTO)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при обновлении иконки: %w", err))
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
// @Failure 400 {object} errors.ErrorResponse "Неверный формат ID"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/manage/icons/{id} [delete]
func (h *IconHandler) DeleteIcon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id", "неверный формат ID"))
		return
	}

	err = h.service.DeleteIcon(c.Request.Context(), uint(id))
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при удалении иконки: %w", err))
		return
	}

	c.Status(http.StatusNoContent)
}

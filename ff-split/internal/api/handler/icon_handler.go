package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
)

// GetIcons возвращает список иконок
func (s *ServerHandler) GetIcons(c *gin.Context) {
	icons, err := s.iconService.GetIcons(c.Request.Context())
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении иконок: %w", err))
		return
	}

	apiIcons := make([]api.IconDTO, 0, len(icons))
	for _, icon := range icons {
		apiIcons = append(apiIcons, convertIconToAPI(&icon))
	}

	c.JSON(http.StatusOK, apiIcons)
}

// GetIconByID возвращает иконку по ID
func (s *ServerHandler) GetIconByID(c *gin.Context, id int) {
	icon, err := s.iconService.GetIconByID(c.Request.Context(), uint(id))
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении иконки: %w", err))
		return
	}

	if icon == nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{Error: "иконка не найдена"})
		return
	}

	c.JSON(http.StatusOK, convertIconToAPI(icon))
}

// CreateIcon создает новую иконку
func (s *ServerHandler) CreateIcon(c *gin.Context) {
	var apiRequest api.IconRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "некорректные данные запроса"})
		return
	}

	dtoRequest := dto.IconFullDTO{
		Name:     apiRequest.Name,
		FileUUID: apiRequest.FileUuid,
	}

	icon, err := s.iconService.CreateIcon(c.Request.Context(), &dtoRequest)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при создании иконки: %w", err))
		return
	}

	c.JSON(http.StatusCreated, convertIconToAPI(icon))
}

// UpdateIcon обновляет иконку
func (s *ServerHandler) UpdateIcon(c *gin.Context, id int) {
	var apiRequest api.IconRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "некорректные данные запроса"})
		return
	}

	dtoRequest := dto.IconFullDTO{
		ID:       uint(id),
		Name:     apiRequest.Name,
		FileUUID: apiRequest.FileUuid,
	}

	icon, err := s.iconService.UpdateIcon(c.Request.Context(), uint(id), &dtoRequest)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при обновлении иконки: %w", err))
		return
	}

	c.JSON(http.StatusOK, convertIconToAPI(icon))
}

// DeleteIcon удаляет иконку
func (s *ServerHandler) DeleteIcon(c *gin.Context, id int) {
	err := s.iconService.DeleteIcon(c.Request.Context(), uint(id))
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при удалении иконки: %w", err))
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{Success: true})
}

// Helper functions

func convertIconToAPI(icon *dto.IconFullDTO) api.IconDTO {
	id := int(icon.ID)
	return api.IconDTO{
		Id:           &id,
		Name:         &icon.Name,
		FileUuid:     &icon.FileUUID,
		ExternalUuid: &icon.FileUUID,
	}
}

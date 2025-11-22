package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
)

// GetCategories возвращает список категорий
func (s *ServerHandler) GetCategories(c *gin.Context, params api.GetCategoriesParams) {
	categoryType := string(params.CategoryType)

	categories, err := s.categoryService.GetCategories(c.Request.Context(), categoryType)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении категорий: %w", err))
		return
	}

	apiCategories := make([]api.CategoryResponse, 0, len(categories))
	for _, cat := range categories {
		apiCategories = append(apiCategories, convertCategoryToAPI(&cat))
	}

	c.JSON(http.StatusOK, api.CategoryListResponse{Categories: &apiCategories})
}

// GetCategoryByID возвращает категорию по ID
func (s *ServerHandler) GetCategoryByID(c *gin.Context, id int, params api.GetCategoryByIDParams) {
	categoryType := string(params.CategoryType)

	category, err := s.categoryService.GetCategoryByID(c.Request.Context(), id, categoryType)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении категории: %w", err))
		return
	}

	if category == nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
		Id: c.GetHeader("X-Request-ID"),
		Error: api.ErrorResponseDetail{
			Code:    "not_found",
			Message: "категория не найдена",
		},
	})
		return
	}

	c.JSON(http.StatusOK, convertCategoryToAPI(category))
}

// CreateCategory создает новую категорию
func (s *ServerHandler) CreateCategory(c *gin.Context, params api.CreateCategoryParams) {
	var apiRequest api.CategoryRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
		Id: c.GetHeader("X-Request-ID"),
		Error: api.ErrorResponseDetail{
			Code:    "validation",
			Message: "некорректные данные запроса",
		},
	})
		return
	}

	categoryType := string(params.CategoryType)

	dtoCategory := &service.CategoryDTO{
		Name:   apiRequest.Name,
		IconID: apiRequest.IconId,
	}

	_, err := s.iconService.GetIconByID(c.Request.Context(), uint(apiRequest.IconId))
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении иконки: %w", err))
		return
	}

	category, err := s.categoryService.CreateCategory(c.Request.Context(), dtoCategory, categoryType)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при создании категории: %w", err))
		return
	}

	c.JSON(http.StatusCreated, convertCategoryToAPI(category))
}

// UpdateCategory обновляет категорию
func (s *ServerHandler) UpdateCategory(c *gin.Context, id int, params api.UpdateCategoryParams) {
	var apiRequest api.CategoryRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
		Id: c.GetHeader("X-Request-ID"),
		Error: api.ErrorResponseDetail{
			Code:    "validation",
			Message: "некорректные данные запроса",
		},
	})
		return
	}

	categoryType := string(params.CategoryType)

	dtoCategory := &service.CategoryDTO{
		ID:     id,
		Name:   apiRequest.Name,
		IconID: apiRequest.IconId,
	}

	_, err := s.iconService.GetIconByID(c.Request.Context(), uint(apiRequest.IconId))
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении иконки: %w", err))
		return
	}

	category, err := s.categoryService.UpdateCategory(c.Request.Context(), id, dtoCategory, categoryType)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при обновлении категории: %w", err))
		return
	}

	c.JSON(http.StatusOK, convertCategoryToAPI(category))
}

// DeleteCategory удаляет категорию
func (s *ServerHandler) DeleteCategory(c *gin.Context, id int, params api.DeleteCategoryParams) {
	categoryType := string(params.CategoryType)

	err := s.categoryService.DeleteCategory(c.Request.Context(), id, categoryType)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при удалении категории: %w", err))
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{Success: true})
}

// Helper functions

func convertCategoryToAPI(cat *service.CategoryDTO) api.CategoryResponse {
	var icon *api.IconDTO
	apiIcon := api.IconDTO{
		Id:           &cat.Icon.ID,
		Name:         &cat.Icon.Name,
		ExternalUuid: &cat.Icon.ExternalUuid,
	}
	icon = &apiIcon

	return api.CategoryResponse{
		Id:   &cat.ID,
		Name: &cat.Name,
		Icon: icon,
	}
}

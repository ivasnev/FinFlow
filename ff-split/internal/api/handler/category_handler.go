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
func (s *ServerHandler) GetCategories(c *gin.Context) {
	// TODO: добавить category_type как query параметр в OpenAPI
	categoryType := c.Query("category_type")
	if categoryType == "" {
		categoryType = "transaction" // default type
	}

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
func (s *ServerHandler) GetCategoryByID(c *gin.Context, id int) {
	// TODO: добавить category_type как query параметр в OpenAPI
	categoryType := c.Query("category_type")
	if categoryType == "" {
		categoryType = "transaction" // default type
	}

	category, err := s.categoryService.GetCategoryByID(c.Request.Context(), id, categoryType)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении категории: %w", err))
		return
	}

	if category == nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{Error: "категория не найдена"})
		return
	}

	c.JSON(http.StatusOK, convertCategoryToAPI(category))
}

// CreateCategory создает новую категорию
func (s *ServerHandler) CreateCategory(c *gin.Context) {
	var apiRequest api.CategoryRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "некорректные данные запроса"})
		return
	}

	// TODO: добавить category_type как query параметр в OpenAPI
	categoryType := c.Query("category_type")
	if categoryType == "" {
		categoryType = "transaction" // default type
	}

	var iconID int
	if apiRequest.IconId != nil {
		iconID = *apiRequest.IconId
	}

	dtoCategory := &service.CategoryDTO{
		Name:   apiRequest.Name,
		IconID: iconID,
	}

	category, err := s.categoryService.CreateCategory(c.Request.Context(), dtoCategory, categoryType)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при создании категории: %w", err))
		return
	}

	c.JSON(http.StatusCreated, convertCategoryToAPI(category))
}

// UpdateCategory обновляет категорию
func (s *ServerHandler) UpdateCategory(c *gin.Context, id int) {
	var apiRequest api.CategoryRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "некорректные данные запроса"})
		return
	}

	// TODO: добавить category_type как query параметр в OpenAPI
	categoryType := c.Query("category_type")
	if categoryType == "" {
		categoryType = "transaction" // default type
	}

	var iconID int
	if apiRequest.IconId != nil {
		iconID = *apiRequest.IconId
	}

	dtoCategory := &service.CategoryDTO{
		ID:     id,
		Name:   apiRequest.Name,
		IconID: iconID,
	}

	category, err := s.categoryService.UpdateCategory(c.Request.Context(), id, dtoCategory, categoryType)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при обновлении категории: %w", err))
		return
	}

	c.JSON(http.StatusOK, convertCategoryToAPI(category))
}

// DeleteCategory удаляет категорию
func (s *ServerHandler) DeleteCategory(c *gin.Context, id int) {
	// TODO: добавить category_type как query параметр в OpenAPI
	categoryType := c.Query("category_type")
	if categoryType == "" {
		categoryType = "transaction" // default type
	}

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

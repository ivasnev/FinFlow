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

// CategoryHandler реализует интерфейс handler.CategoryHandlerInterface
type CategoryHandler struct {
	service service.CategoryServiceInterface
}

// NewCategoryHandler создает новый экземпляр CategoryHandlerInterface
func NewCategoryHandler(service service.CategoryServiceInterface) *CategoryHandler {
	return &CategoryHandler{
		service: service,
	}
}

// Options обрабатывает запрос на получение списка доступных типов категорий
// @Summary Получить типы категорий
// @Description Возвращает список всех доступных типов категорий
// @Tags категории
// @Accept json
// @Produce json
// @Success 200 {array} string "Список типов категорий"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/category [options]
func (h *CategoryHandler) Options(c *gin.Context) {
	types, err := h.service.GetCategoryTypes()
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении типов категорий: %w", err))
		return
	}

	c.JSON(http.StatusOK, types)
}

// GetCategories обрабатывает запрос на получение списка категорий
// @Summary Получить категории
// @Description Возвращает список категорий указанного типа
// @Tags категории
// @Accept json
// @Produce json
// @Param category_type query string true "Тип категории"
// @Success 200 {object} dto.CategoryListResponse "Список категорий"
// @Failure 400 {object} errors.ErrorResponse "Не указан тип категорий"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/category [get]
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем тип категории из query параметра
	categoryType := c.Query("category_type")
	if categoryType == "" {
		errors.HTTPErrorHandler(c, errors.NewValidationError("category_type", "не указан тип категорий"))
		return
	}

	categories, err := h.service.GetCategories(ctx, categoryType)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении категорий: %w", err))
		return
	}

	// Преобразуем модели в DTO для ответа
	response := dto.CategoryListResponse{
		Categories: make([]dto.CategoryResponse, 0, len(categories)),
	}

	for _, category := range categories {
		response.Categories = append(response.Categories, dto.CategoryResponse{
			ID:   category.ID,
			Name: category.Name,
			Icon: category.Icon,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetCategoryByID обрабатывает запрос на получение категории по ID
// @Summary Получить категорию по ID
// @Description Возвращает информацию о конкретной категории по её ID
// @Tags категории
// @Accept json
// @Produce json
// @Param id path int true "ID категории"
// @Param category_type query string true "Тип категории"
// @Success 200 {object} dto.CategoryResponse "Информация о категории"
// @Failure 400 {object} errors.ErrorResponse "Некорректный ID категории или не указан тип"
// @Failure 404 {object} errors.ErrorResponse "Категория не найдена"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/category/{id} [get]
func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID категории из URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id", "некорректный ID категории"))
		return
	}

	// Получаем тип категории из query параметра
	categoryType := c.Query("category_type")
	if categoryType == "" {
		errors.HTTPErrorHandler(c, errors.NewValidationError("category_type", "не указан тип категорий"))
		return
	}

	category, err := h.service.GetCategoryByID(ctx, id, categoryType)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении категории: %w", err))
		return
	}

	if category == nil {
		errors.HTTPErrorHandler(c, errors.NewEntityNotFoundError(idStr, "категория"))
		return
	}

	c.JSON(http.StatusOK, dto.CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
		Icon: category.Icon,
	})
}

// CreateCategory обрабатывает запрос на создание новой категории
// @Summary Создать новую категорию
// @Description Создает новую категорию указанного типа
// @Tags категории
// @Accept json
// @Produce json
// @Param category_type query string true "Тип категории"
// @Param category body dto.CategoryRequest true "Данные категории"
// @Success 201 {object} dto.CategoryResponse "Созданная категория"
// @Failure 400 {object} errors.ErrorResponse "Некорректные данные запроса или не указан тип"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/manage/category [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем данные запроса
	var request dto.CategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("request_body", err.Error()))
		return
	}

	// Получаем тип категории из query параметра
	categoryType := c.Query("category_type")
	if categoryType == "" {
		errors.HTTPErrorHandler(c, errors.NewValidationError("category_type", "не указан тип категорий"))
		return
	}

	category := &dto.CategoryDTO{
		Name:   request.Name,
		IconID: request.IconID,
	}

	createdCategory, err := h.service.CreateCategory(ctx, category, categoryType)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при создании категории: %w", err))
		return
	}

	c.JSON(http.StatusCreated, dto.CategoryResponse{
		ID:   createdCategory.ID,
		Name: createdCategory.Name,
		Icon: createdCategory.Icon,
	})
}

// UpdateCategory обрабатывает запрос на обновление категории
// @Summary Обновить категорию
// @Description Обновляет существующую категорию по ID
// @Tags категории
// @Accept json
// @Produce json
// @Param id path int true "ID категории"
// @Param category_type query string true "Тип категории"
// @Param category body dto.CategoryRequest true "Данные категории"
// @Success 200 {object} dto.CategoryResponse "Обновленная категория"
// @Failure 400 {object} errors.ErrorResponse "Некорректные данные запроса или не указан тип"
// @Failure 404 {object} errors.ErrorResponse "Категория не найдена"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/manage/category/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID категории из URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id", "некорректный ID категории"))
		return
	}

	// Получаем данные запроса
	var request dto.CategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("request_body", err.Error()))
		return
	}

	// Получаем тип категории из query параметра
	categoryType := c.Query("category_type")
	if categoryType == "" {
		errors.HTTPErrorHandler(c, errors.NewValidationError("category_type", "не указан тип категорий"))
		return
	}

	category := &dto.CategoryDTO{
		Name:   request.Name,
		IconID: request.IconID,
	}

	updatedCategory, err := h.service.UpdateCategory(ctx, id, category, categoryType)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при обновлении категории: %w", err))
		return
	}

	if updatedCategory == nil {
		errors.HTTPErrorHandler(c, errors.NewEntityNotFoundError(idStr, "категория"))
		return
	}

	c.JSON(http.StatusOK, dto.CategoryResponse{
		ID:   updatedCategory.ID,
		Name: updatedCategory.Name,
		Icon: updatedCategory.Icon,
	})
}

// DeleteCategory обрабатывает запрос на удаление категории
// @Summary Удалить категорию
// @Description Удаляет категорию по ID
// @Tags категории
// @Accept json
// @Produce json
// @Param id path int true "ID категории"
// @Param category_type query string true "Тип категории"
// @Success 200 {object} dto.SuccessResponse "Категория успешно удалена"
// @Failure 400 {object} errors.ErrorResponse "Некорректный ID категории или не указан тип"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/manage/category/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID категории из URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id", "некорректный ID категории"))
		return
	}

	// Получаем тип категории из query параметра
	categoryType := c.Query("category_type")
	if categoryType == "" {
		errors.HTTPErrorHandler(c, errors.NewValidationError("category_type", "не указан тип категорий"))
		return
	}

	if err := h.service.DeleteCategory(ctx, id, categoryType); err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при удалении категории: %w", err))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
	})
}

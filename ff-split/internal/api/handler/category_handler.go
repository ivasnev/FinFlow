package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
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
func (h *CategoryHandler) Options(c *gin.Context) {
	types, err := h.service.GetCategoryTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при получении типов категорий: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, types)
}

// GetCategories обрабатывает запрос на получение списка категорий
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем тип категории из query параметра
	categoryType := c.Query("category_type")
	if categoryType == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Ошибка при получении категорий: не указан тип категорий",
		})
		return
	}

	categories, err := h.service.GetCategories(ctx, categoryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при получении категорий: " + err.Error(),
		})
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
func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID категории из URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный ID категории",
		})
		return
	}

	// Получаем тип категории из query параметра
	categoryType := c.Query("category_type")
	if categoryType == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Ошибка при получении категорий: не указан тип категорий",
		})
		return
	}

	category, err := h.service.GetCategoryByID(ctx, id, categoryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при получении категории: " + err.Error(),
		})
		return
	}

	if category == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Категория не найдена",
		})
		return
	}

	c.JSON(http.StatusOK, dto.CategoryResponse{
		ID:   category.ID,
		Name: category.Name,
		Icon: category.Icon,
	})
}

// CreateCategory обрабатывает запрос на создание новой категории
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем данные запроса
	var request dto.CategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректные данные запроса: " + err.Error(),
		})
		return
	}

	// Получаем тип категории из query параметра
	categoryType := c.Query("category_type")
	if categoryType == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Ошибка при получении категорий: не указан тип категорий",
		})
		return
	}

	category := &dto.CategoryDTO{
		Name:   request.Name,
		IconID: request.IconID,
	}

	createdCategory, err := h.service.CreateCategory(ctx, category, categoryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при создании категории: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.CategoryResponse{
		ID:   createdCategory.ID,
		Name: createdCategory.Name,
		Icon: createdCategory.Icon,
	})
}

// UpdateCategory обрабатывает запрос на обновление категории
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID категории из URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный ID категории",
		})
		return
	}

	// Получаем данные запроса
	var request dto.CategoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректные данные запроса: " + err.Error(),
		})
		return
	}

	// Получаем тип категории из query параметра
	categoryType := c.Query("category_type")
	if categoryType == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Ошибка при получении категорий: не указан тип категорий",
		})
		return
	}

	category := &dto.CategoryDTO{
		Name:   request.Name,
		IconID: request.IconID,
	}

	updatedCategory, err := h.service.UpdateCategory(ctx, id, category, categoryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при обновлении категории: " + err.Error(),
		})
		return
	}

	if updatedCategory == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Категория не найдена",
		})
		return
	}

	c.JSON(http.StatusOK, dto.CategoryResponse{
		ID:   updatedCategory.ID,
		Name: updatedCategory.Name,
		Icon: updatedCategory.Icon,
	})
}

// DeleteCategory обрабатывает запрос на удаление категории
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID категории из URL
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный ID категории",
		})
		return
	}

	// Получаем тип категории из query параметра
	categoryType := c.Query("category_type")
	if categoryType == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Ошибка при получении категорий: не указан тип категорий",
		})
		return
	}

	if err := h.service.DeleteCategory(ctx, id, categoryType); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при удалении категории: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
	})
}

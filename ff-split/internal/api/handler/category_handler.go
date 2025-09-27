package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// CategoryHandler обработчик запросов для категорий
type CategoryHandler struct {
	categoryService service.CategoryService
}

// NewCategoryHandler создает новый экземпляр обработчика категорий
func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// GetCategories обрабатывает запрос на получение всех категорий
// @Summary Получить все категории
// @Description Получить список всех доступных категорий
// @Tags Category
// @Accept json
// @Produce json
// @Success 200 {array} models.CategoryResponse
// @Router /category [get]
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	categories, err := h.categoryService.GetCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategoryByID обрабатывает запрос на получение категории по ID
// @Summary Получить категорию по ID
// @Description Получить информацию о категории по её ID
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "ID категории"
// @Success 200 {object} models.CategoryResponse
// @Failure 404 {object} map[string]string
// @Router /category/{id} [get]
func (h *CategoryHandler) GetCategoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID"})
		return
	}

	category, err := h.categoryService.GetCategoryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "категория не найдена"})
		return
	}

	c.JSON(http.StatusOK, category)
}

// CreateCategory обрабатывает запрос на создание новой категории
// @Summary Создать новую категорию
// @Description Создать новую категорию с указанными данными
// @Tags Category
// @Accept json
// @Produce json
// @Param category body models.Category true "Данные категории"
// @Success 201 {object} models.CategoryResponse
// @Failure 400 {object} map[string]string
// @Router /category [post]
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdCategory, err := h.categoryService.CreateCategory(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdCategory)
}

// UpdateCategory обрабатывает запрос на обновление категории
// @Summary Обновить категорию
// @Description Обновить данные категории по её ID
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "ID категории"
// @Param category body models.Category true "Данные категории"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /category/{id} [put]
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID"})
		return
	}

	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.ID = id

	if err := h.categoryService.UpdateCategory(c.Request.Context(), category); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "категория успешно обновлена"})
}

// DeleteCategory обрабатывает запрос на удаление категории
// @Summary Удалить категорию
// @Description Удалить категорию по её ID
// @Tags Category
// @Accept json
// @Produce json
// @Param id path int true "ID категории"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /category/{id} [delete]
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID"})
		return
	}

	if err := h.categoryService.DeleteCategory(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "категория успешно удалена"})
}

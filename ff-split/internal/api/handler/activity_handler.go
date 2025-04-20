package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// ActivityHandler обработчик запросов для активностей
type ActivityHandler struct {
	activityService service.ActivityService
}

// NewActivityHandler создает новый экземпляр обработчика активностей
func NewActivityHandler(activityService service.ActivityService) *ActivityHandler {
	return &ActivityHandler{
		activityService: activityService,
	}
}

// GetActivitiesByEventID обрабатывает запрос на получение всех активностей мероприятия
// @Summary Получить все активности мероприятия
// @Description Получить список всех активностей определенного мероприятия
// @Tags Activity
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Success 200 {array} models.ActivityResponse
// @Failure 400 {object} map[string]string
// @Router /event/{id_event}/activity [get]
func (h *ActivityHandler) GetActivitiesByEventID(c *gin.Context) {
	eventIDStr := c.Param("id_event")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID мероприятия"})
		return
	}

	activities, err := h.activityService.GetActivitiesByEventID(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, activities)
}

// GetActivityByID обрабатывает запрос на получение активности по ID
// @Summary Получить активность по ID
// @Description Получить информацию об активности по её ID
// @Tags Activity
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id path int true "ID активности"
// @Success 200 {object} models.ActivityResponse
// @Failure 404 {object} map[string]string
// @Router /event/{id_event}/activity/{id} [get]
func (h *ActivityHandler) GetActivityByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID активности"})
		return
	}

	activity, err := h.activityService.GetActivityByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "активность не найдена"})
		return
	}

	c.JSON(http.StatusOK, activity)
}

// CreateActivity обрабатывает запрос на создание новой активности
// @Summary Создать новую активность
// @Description Создать новую активность для определенного мероприятия
// @Tags Activity
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param activity body models.Activity true "Данные активности"
// @Success 201 {object} models.ActivityResponse
// @Failure 400 {object} map[string]string
// @Router /event/{id_event}/activity [post]
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	eventIDStr := c.Param("id_event")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID мероприятия"})
		return
	}

	var activity models.Activity
	if err := c.ShouldBindJSON(&activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	activity.IDEvent = eventID

	createdActivity, err := h.activityService.CreateActivity(c.Request.Context(), activity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdActivity)
}

// UpdateActivity обрабатывает запрос на обновление активности
// @Summary Обновить активность
// @Description Обновить данные активности по её ID
// @Tags Activity
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id path int true "ID активности"
// @Param activity body models.Activity true "Данные активности"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /event/{id_event}/activity/{id} [put]
func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID активности"})
		return
	}

	eventIDStr := c.Param("id_event")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID мероприятия"})
		return
	}

	var activity models.Activity
	if err := c.ShouldBindJSON(&activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	activity.ID = id
	activity.IDEvent = eventID

	if err := h.activityService.UpdateActivity(c.Request.Context(), activity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "активность успешно обновлена"})
}

// DeleteActivity обрабатывает запрос на удаление активности
// @Summary Удалить активность
// @Description Удалить активность по её ID
// @Tags Activity
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id path int true "ID активности"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /event/{id_event}/activity/{id} [delete]
func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID активности"})
		return
	}

	if err := h.activityService.DeleteActivity(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "активность успешно удалена"})
}

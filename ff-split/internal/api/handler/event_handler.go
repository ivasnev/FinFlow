package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// EventHandler обработчик запросов для мероприятий
type EventHandler struct {
	eventService service.EventService
}

// NewEventHandler создает новый экземпляр обработчика мероприятий
func NewEventHandler(eventService service.EventService) *EventHandler {
	return &EventHandler{
		eventService: eventService,
	}
}

// GetEvents обрабатывает запрос на получение всех мероприятий
// @Summary Получить все мероприятия
// @Description Получить список всех доступных мероприятий
// @Tags Event
// @Accept json
// @Produce json
// @Success 200 {array} models.EventResponse
// @Router /events [get]
func (h *EventHandler) GetEvents(c *gin.Context) {
	events, err := h.eventService.GetEvents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

// GetEventByID обрабатывает запрос на получение мероприятия по ID
// @Summary Получить мероприятие по ID
// @Description Получить информацию о мероприятии по его ID
// @Tags Event
// @Accept json
// @Produce json
// @Param id path int true "ID мероприятия"
// @Success 200 {object} models.EventResponse
// @Failure 404 {object} map[string]string
// @Router /event/{id} [get]
func (h *EventHandler) GetEventByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID"})
		return
	}

	event, err := h.eventService.GetEventByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "мероприятие не найдено"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// CreateEvent обрабатывает запрос на создание нового мероприятия
// @Summary Создать новое мероприятие
// @Description Создать новое мероприятие с указанными данными
// @Tags Event
// @Accept json
// @Produce json
// @Param event body models.EventRequest true "Данные мероприятия"
// @Success 201 {object} models.EventResponse
// @Failure 400 {object} map[string]string
// @Router /event [post]
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var eventRequest models.EventRequest
	if err := c.ShouldBindJSON(&eventRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdEvent, err := h.eventService.CreateEvent(c.Request.Context(), eventRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdEvent)
}

// UpdateEvent обрабатывает запрос на обновление мероприятия
// @Summary Обновить мероприятие
// @Description Обновить данные мероприятия по его ID
// @Tags Event
// @Accept json
// @Produce json
// @Param id path int true "ID мероприятия"
// @Param event body models.EventRequest true "Данные мероприятия"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /event/{id} [put]
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID"})
		return
	}

	var eventRequest models.EventRequest
	if err := c.ShouldBindJSON(&eventRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.eventService.UpdateEvent(c.Request.Context(), id, eventRequest); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "мероприятие успешно обновлено"})
}

// DeleteEvent обрабатывает запрос на удаление мероприятия
// @Summary Удалить мероприятие
// @Description Удалить мероприятие по его ID
// @Tags Event
// @Accept json
// @Produce json
// @Param id path int true "ID мероприятия"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /event/{id} [delete]
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный ID"})
		return
	}

	if err := h.eventService.DeleteEvent(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "мероприятие успешно удалено"})
}

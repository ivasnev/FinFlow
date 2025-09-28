package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// EventHandler реализует интерфейс handler.EventHandlerInterface
type EventHandler struct {
	service service.EventServiceInterface
}

// NewEventHandler создает новый экземпляр EventHandlerInterface
func NewEventHandler(service service.EventServiceInterface) *EventHandler {
	return &EventHandler{
		service: service,
	}
}

// GetEvents обрабатывает запрос на получение списка мероприятий
func (h *EventHandler) GetEvents(c *gin.Context) {
	ctx := c.Request.Context()

	events, err := h.service.GetEvents(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при получении мероприятий: " + err.Error(),
		})
		return
	}

	// Преобразуем модели в DTO для ответа
	response := dto.EventListResponse{
		Events: make([]dto.EventResponse, 0, len(events)),
	}

	for _, event := range events {
		response.Events = append(response.Events, dto.EventResponse{
			ID:          event.ID,
			Name:        event.Name,
			Description: event.Description,
			CategoryID:  event.CategoryID,
			ImageID:     event.ImageID,
			Status:      event.Status,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetEventByID обрабатывает запрос на получение мероприятия по ID
func (h *EventHandler) GetEventByID(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID мероприятия из URL
	idStr := c.Param("id_event")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный ID мероприятия",
		})
		return
	}

	event, err := h.service.GetEventByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при получении мероприятия: " + err.Error(),
		})
		return
	}

	if event == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Мероприятие не найдено",
		})
		return
	}

	c.JSON(http.StatusOK, dto.EventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		CategoryID:  event.CategoryID,
		ImageID:     event.ImageID,
		Status:      event.Status,
	})
}

// CreateEvent обрабатывает запрос на создание нового мероприятия
func (h *EventHandler) CreateEvent(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем данные запроса
	var request dto.EventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректные данные запроса: " + err.Error(),
		})
		return
	}

	// Преобразуем DTO в модель
	event := &models.Event{
		Name:        request.Name,
		Description: request.Description,
		CategoryID:  request.CategoryID,
		ImageID:     request.ImageID,
		Status:      request.Status,
	}

	if event.Status == "" {
		event.Status = "active" // Статус по умолчанию
	}

	createdEvent, err := h.service.CreateEvent(ctx, event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при создании мероприятия: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, dto.EventResponse{
		ID:          createdEvent.ID,
		Name:        createdEvent.Name,
		Description: createdEvent.Description,
		CategoryID:  createdEvent.CategoryID,
		ImageID:     createdEvent.ImageID,
		Status:      createdEvent.Status,
	})
}

// UpdateEvent обрабатывает запрос на обновление мероприятия
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID мероприятия из URL
	idStr := c.Param("id_event")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный ID мероприятия",
		})
		return
	}

	// Получаем данные запроса
	var request dto.EventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректные данные запроса: " + err.Error(),
		})
		return
	}

	// Преобразуем DTO в модель
	event := &models.Event{
		Name:        request.Name,
		Description: request.Description,
		CategoryID:  request.CategoryID,
		ImageID:     request.ImageID,
		Status:      request.Status,
	}

	updatedEvent, err := h.service.UpdateEvent(ctx, id, event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при обновлении мероприятия: " + err.Error(),
		})
		return
	}

	if updatedEvent == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Мероприятие не найдено",
		})
		return
	}

	c.JSON(http.StatusOK, dto.EventResponse{
		ID:          updatedEvent.ID,
		Name:        updatedEvent.Name,
		Description: updatedEvent.Description,
		CategoryID:  updatedEvent.CategoryID,
		ImageID:     updatedEvent.ImageID,
		Status:      updatedEvent.Status,
	})
}

// DeleteEvent обрабатывает запрос на удаление мероприятия
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID мероприятия из URL
	idStr := c.Param("id_event")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный ID мероприятия",
		})
		return
	}

	if err := h.service.DeleteEvent(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при удалении мероприятия: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
	})
}

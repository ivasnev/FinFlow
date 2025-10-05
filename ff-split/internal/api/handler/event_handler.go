package handler

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// EventHandler реализует интерфейс handler.EventHandlerInterface
type EventHandler struct {
	service     service.EventServiceInterface
	userService service.UserServiceInterface
}

// NewEventHandler создает новый экземпляр EventHandlerInterface
func NewEventHandler(service service.EventServiceInterface, userService service.UserServiceInterface) *EventHandler {
	return &EventHandler{
		service:     service,
		userService: userService,
	}
}

// GetEvents обрабатывает запрос на получение списка мероприятий
// @Summary Получить список мероприятий
// @Description Возвращает список всех мероприятий
// @Tags мероприятия
// @Accept json
// @Produce json
// @Success 200 {object} dto.EventListResponse "Список мероприятий"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event [get]
func (h *EventHandler) GetEvents(c *gin.Context) {
	ctx := c.Request.Context()

	events, err := h.service.GetEvents(ctx)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении мероприятий: %w", err))
		return
	}

	// Преобразуем модели в DTO для ответа
	response := dto.EventListResponse{
		Events: make([]dto.EventResponse, 0, len(events)),
	}

	for _, event := range events {
		// Заглушка для баланса
		var balance *int = nil
		// Здесь будет расчет баланса в будущем

		response.Events = append(response.Events, dto.EventResponse{
			ID:          event.ID,
			Name:        event.Name,
			Description: event.Description,
			CategoryID:  event.CategoryID,
			PhotoID:     event.ImageID,
			Balance:     balance,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetEventByID обрабатывает запрос на получение мероприятия по ID
// @Summary Получить мероприятие по ID
// @Description Возвращает информацию о конкретном мероприятии по его ID
// @Tags мероприятия
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Success 200 {object} dto.EventResponse "Информация о мероприятии"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат ID мероприятия"
// @Failure 404 {object} errors.ErrorResponse "Мероприятие не найдено"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event} [get]
func (h *EventHandler) GetEventByID(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID мероприятия из URL
	idStr := c.Param("id_event")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_event", "некорректный ID мероприятия"))
		return
	}

	event, err := h.service.GetEventByID(ctx, id)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении мероприятия: %w", err))
		return
	}

	if event == nil {
		errors.HTTPErrorHandler(c, errors.NewEntityNotFoundError(idStr, "event"))
		return
	}

	// Заглушка для баланса
	var balance *int = nil
	// Здесь будет расчет баланса в будущем

	c.JSON(http.StatusOK, dto.EventResponse{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		CategoryID:  event.CategoryID,
		PhotoID:     event.ImageID,
		Balance:     balance,
	})
}

// CreateEvent обрабатывает запрос на создание нового мероприятия
// @Summary Создать новое мероприятие
// @Description Создает новое мероприятие с указанным пользователем в качестве создателя
// @Tags мероприятия
// @Accept json
// @Produce json
// @Param id_user path int true "ID пользователя (создателя)"
// @Param event body dto.EventRequest true "Данные мероприятия"
// @Success 201 {object} dto.EventResponse "Созданное мероприятие"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат данных запроса"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/ [post]
func (h *EventHandler) CreateEvent(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем данные запроса
	var request dto.EventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("event", "некорректные данные запроса"))
		return
	}
	if rawID, ok := c.Get("user_id"); ok {
		if idInt, ok := rawID.(int64); ok && !slices.Contains(request.Members.UserIDs, idInt) {
			request.Members.UserIDs = append(request.Members.UserIDs, idInt)
		}
	}

	event, err := h.service.CreateEvent(ctx, &request)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка создания ивента: %w", err))
		return
	}
	c.JSON(http.StatusOK, event)
}

// UpdateEvent обрабатывает запрос на обновление мероприятия
// @Summary Обновить мероприятие
// @Description Обновляет существующее мероприятие по ID
// @Tags мероприятия
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param event body dto.EventRequest true "Данные мероприятия"
// @Success 200 {object} dto.EventResponse "Обновленное мероприятие"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат данных запроса"
// @Failure 404 {object} errors.ErrorResponse "Мероприятие не найдено"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event} [put]
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID мероприятия из URL
	idStr := c.Param("id_event")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_event", "некорректный ID мероприятия"))
		return
	}

	// Получаем данные запроса
	var request dto.EventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("request_body", "некорректные данные запроса"))
		return
	}

	event, err := h.service.UpdateEvent(ctx, id, &request)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка обновления мероприятия: %w", err))
		return
	}
	c.JSON(http.StatusOK, event)
}

// DeleteEvent обрабатывает запрос на удаление мероприятия
// @Summary Удалить мероприятие
// @Description Удаляет мероприятие по ID
// @Tags мероприятия
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Success 204 "Мероприятие успешно удалено"
// @Failure 400 {object} errors.ErrorResponse "Неверный формат ID мероприятия"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event} [delete]
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID мероприятия из URL
	idStr := c.Param("id_event")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_event", "некорректный ID мероприятия"))
		return
	}

	if err := h.service.DeleteEvent(ctx, id); err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при удалении мероприятия: %w", err))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
	})
}

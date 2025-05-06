package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// ActivityHandler содержит методы для работы с логами активности
type ActivityHandler struct {
	service service.ActivityServiceInterface
}

// NewActivityHandler создает новый объект хэндлера активности
func NewActivityHandler(service service.ActivityServiceInterface) *ActivityHandler {
	return &ActivityHandler{
		service: service,
	}
}

// GetActivitiesByEventID возвращает активности для указанного события
// @Summary Получить активности события
// @Description Возвращает список активностей для указанного события
// @Tags активности
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Success 200 {object} dto.ActivityListResponse "Список активностей"
// @Failure 400 {object} errors.ErrorResponse "Некорректный ID события"
// @Failure 404 {object} errors.ErrorResponse "Событие не найдено"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/activity [get]
func (h *ActivityHandler) GetActivitiesByEventID(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID мероприятия из URL
	eventIDStr := c.Param("id_event")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_event", "некорректный ID события"))
		return
	}

	activities, err := h.service.GetActivitiesByEventID(ctx, eventID)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении активностей: %w", err))
		return
	}

	// Преобразуем модели в DTO
	response := dto.ActivityListResponse{
		Activities: make([]dto.ActivityResponse, 0, len(activities)),
	}

	for _, activity := range activities {
		response.Activities = append(response.Activities, dto.ActivityResponse{
			ActivityID:  activity.ID,
			Description: activity.Description,
			IconID:      activity.IconID,
			Datetime:    activity.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetActivityByID обрабатывает запрос на получение активности по ID
// @Summary Получить активность по ID
// @Description Возвращает информацию о конкретной активности по её ID
// @Tags активности
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id_activity path int true "ID активности"
// @Success 200 {object} dto.ActivityResponse "Информация об активности"
// @Failure 400 {object} errors.ErrorResponse "Некорректный ID активности"
// @Failure 404 {object} errors.ErrorResponse "Активность не найдена"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/activity/{id_activity} [get]
func (h *ActivityHandler) GetActivityByID(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID активности из URL
	idStr := c.Param("id_activity")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_activity", "некорректный ID активности"))
		return
	}

	activity, err := h.service.GetActivityByID(ctx, id)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении активности: %w", err))
		return
	}

	if activity == nil {
		errors.HTTPErrorHandler(c, errors.NewEntityNotFoundError(idStr, "активность"))
		return
	}

	c.JSON(http.StatusOK, dto.ActivityResponse{
		ActivityID:  activity.ID,
		Description: activity.Description,
		IconID:      activity.IconID,
		Datetime:    activity.CreatedAt,
	})
}

// CreateActivity создает новую запись об активности
// @Summary Создать активность
// @Description Создает новую запись об активности в событии
// @Tags активности
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param activity body dto.ActivityRequest true "Данные активности"
// @Success 201 {object} dto.ActivityResponse "Созданная активность"
// @Failure 400 {object} errors.ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} errors.ErrorResponse "Событие не найдено"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/activity [post]
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID мероприятия из URL
	eventIDStr := c.Param("id_event")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_event", "некорректный ID события"))
		return
	}

	// Получаем данные активности из тела запроса
	var request dto.ActivityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("request_body", err.Error()))
		return
	}

	// Преобразуем DTO в модель для сервиса
	activity := &models.Activity{
		EventID:     &eventID,
		UserID:      request.UserID,
		Description: request.Description,
		IconID:      request.IconID,
		CreatedAt:   time.Now(), // Устанавливаем текущее время
	}

	// Вызываем сервис для создания активности
	createdActivity, err := h.service.CreateActivity(ctx, activity)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при создании активности: %w", err))
		return
	}

	// Возвращаем созданную активность
	c.JSON(http.StatusCreated, dto.ActivityResponse{
		ActivityID:  createdActivity.ID,
		Description: createdActivity.Description,
		IconID:      request.IconID, // Используем значение из запроса
		Datetime:    createdActivity.CreatedAt,
	})
}

// UpdateActivity обрабатывает запрос на обновление активности
// @Summary Обновить активность
// @Description Обновляет существующую активность по ID
// @Tags активности
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id_activity path int true "ID активности"
// @Param activity body dto.ActivityRequest true "Данные активности"
// @Success 200 {object} dto.ActivityResponse "Обновленная активность"
// @Failure 400 {object} errors.ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} errors.ErrorResponse "Активность не найдена"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/activity/{id_activity} [put]
func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID активности из URL
	idStr := c.Param("id_activity")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_activity", "некорректный ID активности"))
		return
	}

	// Получаем ID события из URL
	eventIDStr := c.Param("id_event")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_event", "некорректный ID события"))
		return
	}

	// Получаем данные запроса
	var request dto.ActivityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("request_body", err.Error()))
		return
	}

	// Преобразуем DTO в модель
	activity := &models.Activity{
		EventID:     &eventID,
		UserID:      request.UserID,
		Description: request.Description,
	}

	updatedActivity, err := h.service.UpdateActivity(ctx, id, activity)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при обновлении активности: %w", err))
		return
	}

	if updatedActivity == nil {
		errors.HTTPErrorHandler(c, errors.NewEntityNotFoundError(idStr, "активность"))
		return
	}

	c.JSON(http.StatusOK, dto.ActivityResponse{
		ActivityID:  updatedActivity.ID,
		Description: updatedActivity.Description,
		IconID:      request.IconID, // Используем значение из запроса
		Datetime:    updatedActivity.CreatedAt,
	})
}

// DeleteActivity обрабатывает запрос на удаление активности
// @Summary Удалить активность
// @Description Удаляет активность по ID
// @Tags активности
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id_activity path int true "ID активности"
// @Success 200 {object} dto.SuccessResponse "Активность успешно удалена"
// @Failure 400 {object} errors.ErrorResponse "Некорректный ID активности"
// @Failure 500 {object} errors.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/activity/{id_activity} [delete]
func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID активности из URL
	idStr := c.Param("id_activity")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("id_activity", "некорректный ID активности"))
		return
	}

	if err := h.service.DeleteActivity(ctx, id); err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при удалении активности: %w", err))
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
	})
}

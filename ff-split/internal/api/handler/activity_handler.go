package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// ActivityHandler реализует интерфейс handler.ActivityHandlerInterface
type ActivityHandler struct {
	service service.ActivityServiceInterface
}

// NewActivityHandler создает новый экземпляр ActivityHandlerInterface
func NewActivityHandler(service service.ActivityServiceInterface) *ActivityHandler {
	return &ActivityHandler{
		service: service,
	}
}

// GetActivitiesByEventID обрабатывает запрос на получение списка активностей по ID мероприятия
// @Summary Получить активности мероприятия
// @Description Возвращает список всех активностей, связанных с указанным мероприятием
// @Tags активности
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Success 200 {object} dto.ActivityListResponse "Список активностей"
// @Failure 400 {object} dto.ErrorResponse "Некорректный ID мероприятия"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/activity [get]
func (h *ActivityHandler) GetActivitiesByEventID(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID мероприятия из URL
	eventIDStr := c.Param("id_event")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный ID мероприятия",
		})
		return
	}

	activities, err := h.service.GetActivitiesByEventID(ctx, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при получении активностей: " + err.Error(),
		})
		return
	}

	// Преобразуем модели в DTO для ответа
	response := dto.ActivityListResponse{
		Activities: make([]dto.ActivityResponse, 0, len(activities)),
	}

	for _, activity := range activities {
		response.Activities = append(response.Activities, dto.ActivityResponse{
			ActivityID:  activity.ID,
			Description: activity.Description,
			IconID:      "", // Пока нет поля в модели, устанавливаем пустое значение
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
// @Failure 400 {object} dto.ErrorResponse "Некорректный ID активности"
// @Failure 404 {object} dto.ErrorResponse "Активность не найдена"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/activity/{id_activity} [get]
func (h *ActivityHandler) GetActivityByID(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID активности из URL
	idStr := c.Param("id_activity")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный ID активности",
		})
		return
	}

	activity, err := h.service.GetActivityByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при получении активности: " + err.Error(),
		})
		return
	}

	if activity == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Активность не найдена",
		})
		return
	}

	c.JSON(http.StatusOK, dto.ActivityResponse{
		ActivityID:  activity.ID,
		Description: activity.Description,
		IconID:      "", // Пока нет поля в модели, устанавливаем пустое значение
		Datetime:    activity.CreatedAt,
	})
}

// CreateActivity обрабатывает запрос на создание новой активности
// @Summary Создать новую активность
// @Description Создает новую активность в рамках указанного мероприятия
// @Tags активности
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param activity body dto.ActivityRequest true "Данные активности"
// @Success 201 {object} dto.ActivityResponse "Созданная активность"
// @Failure 400 {object} dto.ErrorResponse "Некорректные данные запроса"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/activity [post]
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID мероприятия из URL
	eventIDStr := c.Param("id_event")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный ID мероприятия",
		})
		return
	}

	// Получаем данные запроса
	var request dto.ActivityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректные данные запроса: " + err.Error(),
		})
		return
	}

	// Преобразуем DTO в модель
	activity := &models.Activity{
		EventID:     &eventID,
		UserID:      request.UserID,
		Description: request.Description,
		CreatedAt:   time.Now(), // Устанавливаем текущее время
	}

	createdActivity, err := h.service.CreateActivity(ctx, activity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при создании активности: " + err.Error(),
		})
		return
	}

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
// @Failure 400 {object} dto.ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} dto.ErrorResponse "Активность не найдена"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/activity/{id_activity} [put]
func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID активности из URL
	idStr := c.Param("id_activity")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный ID активности",
		})
		return
	}

	// Получаем ID мероприятия из URL
	eventIDStr := c.Param("id_event")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный ID мероприятия",
		})
		return
	}

	// Получаем данные запроса
	var request dto.ActivityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректные данные запроса: " + err.Error(),
		})
		return
	}

	// Преобразуем DTO в модель
	activity := &models.Activity{
		EventID:     &eventID,
		UserID:      request.UserID,
		Description: request.Description,
		// Не обновляем CreatedAt
	}

	updatedActivity, err := h.service.UpdateActivity(ctx, id, activity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при обновлении активности: " + err.Error(),
		})
		return
	}

	if updatedActivity == nil {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: "Активность не найдена",
		})
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
// @Failure 400 {object} dto.ErrorResponse "Некорректный ID активности"
// @Failure 500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/activity/{id_activity} [delete]
func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем ID активности из URL
	idStr := c.Param("id_activity")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Некорректный ID активности",
		})
		return
	}

	if err := h.service.DeleteActivity(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Ошибка при удалении активности: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{
		Success: true,
	})
}

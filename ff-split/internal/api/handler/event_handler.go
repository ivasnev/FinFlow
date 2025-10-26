package handler

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
)

// GetEvents обрабатывает запрос на получение списка мероприятий
func (s *ServerHandler) GetEvents(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем данные пользователя из контекста
	userData, exists := auth.GetUserData(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{
			Error: "пользователь не авторизован",
		})
		return
	}

	// Преобразуем внешний ID во внутренний
	user, err := s.userService.GetUserByExternalUserID(ctx, userData.UserID)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении пользователя: %w", err))
		return
	}

	// Получаем события пользователя с балансами
	serviceEvents, err := s.eventService.GetEventsByUserID(ctx, user.ID)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении мероприятий: %w", err))
		return
	}

	// Преобразуем service типы в API типы
	apiEvents := make([]api.EventResponse, 0, len(serviceEvents))
	for _, event := range serviceEvents {
		apiEvent := api.EventResponse{
			Id:          &event.ID,
			Name:        &event.Name,
			Description: &event.Description,
			CategoryId:  event.CategoryID,
			PhotoId:     &event.PhotoID,
			Balance:     event.Balance,
		}
		apiEvents = append(apiEvents, apiEvent)
	}

	response := api.EventListResponse{
		Events: &apiEvents,
	}

	c.JSON(http.StatusOK, response)
}

// GetEventByID обрабатывает запрос на получение мероприятия по ID
func (s *ServerHandler) GetEventByID(c *gin.Context, idEvent int64) {
	ctx := c.Request.Context()

	event, err := s.eventService.GetEventByID(ctx, idEvent)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении мероприятия: %w", err))
		return
	}

	if event == nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Error: "мероприятие не найдено",
		})
		return
	}

	// Заглушка для баланса
	var balance *int = nil
	// Здесь будет расчет баланса в будущем

	c.JSON(http.StatusOK, api.EventResponse{
		Id:          &event.ID,
		Name:        &event.Name,
		Description: &event.Description,
		CategoryId:  event.CategoryID,
		PhotoId:     &event.ImageID,
		Balance:     balance,
	})
}

// CreateEvent обрабатывает запрос на создание нового мероприятия
func (s *ServerHandler) CreateEvent(c *gin.Context) {
	ctx := c.Request.Context()

	// Получаем данные запроса
	var apiRequest api.EventRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "некорректные данные запроса",
		})
		return
	}

	// Конвертируем API типы в DTO
	dtoRequest := service.EventRequest{
		Name:       apiRequest.Name,
		CategoryID: apiRequest.CategoryId,
	}

	if apiRequest.Description != nil {
		dtoRequest.Description = *apiRequest.Description
	}

	if apiRequest.Members != nil {
		if apiRequest.Members.UserIds != nil {
			dtoRequest.Members.UserIDs = *apiRequest.Members.UserIds
		}
		if apiRequest.Members.DummiesNames != nil {
			dtoRequest.Members.DummiesNames = *apiRequest.Members.DummiesNames
		}
	}

	// Добавляем текущего пользователя к members, если его там нет
	if rawID, ok := c.Get("user_id"); ok {
		if idInt, ok := rawID.(int64); ok && !slices.Contains(dtoRequest.Members.UserIDs, idInt) {
			dtoRequest.Members.UserIDs = append(dtoRequest.Members.UserIDs, idInt)
		}
	}

	eventResponse, err := s.eventService.CreateEvent(ctx, &dtoRequest)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка создания мероприятия: %w", err))
		return
	}

	// Конвертируем DTO ответ в API типы
	apiResponse := api.EventResponse{
		Id:          &eventResponse.ID,
		Name:        &eventResponse.Name,
		Description: &eventResponse.Description,
		CategoryId:  eventResponse.CategoryID,
		PhotoId:     &eventResponse.PhotoID,
		Balance:     eventResponse.Balance,
	}

	c.JSON(http.StatusCreated, apiResponse)
}

// UpdateEvent обрабатывает запрос на обновление мероприятия
func (s *ServerHandler) UpdateEvent(c *gin.Context, idEvent int64) {
	ctx := c.Request.Context()

	// Получаем данные запроса
	var apiRequest api.EventRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "некорректные данные запроса",
		})
		return
	}

	// Конвертируем API типы в DTO
	dtoRequest := service.EventRequest{
		Name:       apiRequest.Name,
		CategoryID: apiRequest.CategoryId,
	}

	if apiRequest.Description != nil {
		dtoRequest.Description = *apiRequest.Description
	}

	if apiRequest.Members != nil {
		if apiRequest.Members.UserIds != nil {
			dtoRequest.Members.UserIDs = *apiRequest.Members.UserIds
		}
		if apiRequest.Members.DummiesNames != nil {
			dtoRequest.Members.DummiesNames = *apiRequest.Members.DummiesNames
		}
	}

	eventResponse, err := s.eventService.UpdateEvent(ctx, idEvent, &dtoRequest)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка обновления мероприятия: %w", err))
		return
	}

	// Конвертируем DTO ответ в API типы
	apiResponse := api.EventResponse{
		Id:          &eventResponse.ID,
		Name:        &eventResponse.Name,
		Description: &eventResponse.Description,
		CategoryId:  eventResponse.CategoryID,
		PhotoId:     &eventResponse.PhotoID,
		Balance:     eventResponse.Balance,
	}

	c.JSON(http.StatusOK, apiResponse)
}

// DeleteEvent обрабатывает запрос на удаление мероприятия
func (s *ServerHandler) DeleteEvent(c *gin.Context, idEvent int64) {
	ctx := c.Request.Context()

	if err := s.eventService.DeleteEvent(ctx, idEvent); err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при удалении мероприятия: %w", err))
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{
		Success: true,
	})
}

package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
)

// GetActivitiesByEventID возвращает активности мероприятия
func (s *ServerHandler) GetActivitiesByEventID(c *gin.Context, idEvent int64) {
	activities, err := s.activityService.GetActivitiesByEventID(c.Request.Context(), idEvent)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении активностей: %w", err))
		return
	}

	apiActivities := make([]api.ActivityResponse, 0, len(activities))
	for _, a := range activities {
		activityID := a.ID
		iconID := a.IconID
		apiActivities = append(apiActivities, api.ActivityResponse{
			ActivityId:  &activityID,
			Description: &a.Description,
			IconId:      &iconID,
			Datetime:    &a.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, api.ActivityListResponse{Activities: &apiActivities})
}

// GetActivityByID возвращает активность по ID
func (s *ServerHandler) GetActivityByID(c *gin.Context, idEvent int64, idActivity int) {
	activity, err := s.activityService.GetActivityByID(c.Request.Context(), idActivity)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении активности: %w", err))
		return
	}

	if activity == nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{
		Id: c.GetHeader("X-Request-ID"),
		Error: api.ErrorResponseDetail{
			Code:    "not_found",
			Message: "активность не найдена",
		},
	})
		return
	}

	activityID := activity.ID
	iconID := activity.IconID
	c.JSON(http.StatusOK, api.ActivityResponse{
		ActivityId:  &activityID,
		Description: &activity.Description,
		IconId:      &iconID,
		Datetime:    &activity.CreatedAt,
	})
}

// CreateActivity создает новую активность
func (s *ServerHandler) CreateActivity(c *gin.Context, idEvent int64) {
	var apiRequest api.ActivityRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
		Id: c.GetHeader("X-Request-ID"),
		Error: api.ErrorResponseDetail{
			Code:    "validation",
			Message: "некорректные данные запроса",
		},
	})
		return
	}

	var iconID int
	if apiRequest.IconId != nil {
		iconID = *apiRequest.IconId
	}

	modelActivity := &models.Activity{
		EventID:     &idEvent,
		Description: apiRequest.Description,
		IconID:      iconID,
		UserID:      apiRequest.UserId,
	}

	activity, err := s.activityService.CreateActivity(c.Request.Context(), modelActivity)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при создании активности: %w", err))
		return
	}

	activityID := activity.ID
	resultIconID := activity.IconID
	c.JSON(http.StatusCreated, api.ActivityResponse{
		ActivityId:  &activityID,
		Description: &activity.Description,
		IconId:      &resultIconID,
		Datetime:    &activity.CreatedAt,
	})
}

// UpdateActivity обновляет активность
func (s *ServerHandler) UpdateActivity(c *gin.Context, idEvent int64, idActivity int) {
	var apiRequest api.ActivityRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
		Id: c.GetHeader("X-Request-ID"),
		Error: api.ErrorResponseDetail{
			Code:    "validation",
			Message: "некорректные данные запроса",
		},
	})
		return
	}

	var iconID int
	if apiRequest.IconId != nil {
		iconID = *apiRequest.IconId
	}

	modelActivity := &models.Activity{
		EventID:     &idEvent,
		Description: apiRequest.Description,
		IconID:      iconID,
		UserID:      apiRequest.UserId,
	}

	activity, err := s.activityService.UpdateActivity(c.Request.Context(), idActivity, modelActivity)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при обновлении активности: %w", err))
		return
	}

	activityID := activity.ID
	resultIconID := activity.IconID
	c.JSON(http.StatusOK, api.ActivityResponse{
		ActivityId:  &activityID,
		Description: &activity.Description,
		IconId:      &resultIconID,
		Datetime:    &activity.CreatedAt,
	})
}

// DeleteActivity удаляет активность
func (s *ServerHandler) DeleteActivity(c *gin.Context, idEvent int64, idActivity int) {
	err := s.activityService.DeleteActivity(c.Request.Context(), idActivity)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при удалении активности: %w", err))
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{Success: true})
}

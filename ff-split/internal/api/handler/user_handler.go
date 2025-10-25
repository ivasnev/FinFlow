package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
)

// GetUserByID возвращает пользователя по ID
func (s *ServerHandler) GetUserByID(c *gin.Context, idUser int64) {
	user, err := s.userService.GetUserByExternalUserID(c.Request.Context(), idUser)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("пользователь не найден: %w", err))
		return
	}

	c.JSON(http.StatusOK, convertUserToProfileAPI(user))
}

// GetUsersByEventID возвращает пользователей мероприятия
func (s *ServerHandler) GetUsersByEventID(c *gin.Context, idEvent int64) {
	users, err := s.userService.GetUsersByEventID(c.Request.Context(), idEvent)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении пользователей: %w", err))
		return
	}

	apiUsers := make([]api.UserProfileDTO, 0, len(users))
	for i := range users {
		apiUsers = append(apiUsers, convertUserToProfileAPI(&users[i]))
	}

	c.JSON(http.StatusOK, api.UserListResponse{Users: &apiUsers})
}

// GetDummiesByEventID возвращает dummy-пользователей мероприятия
func (s *ServerHandler) GetDummiesByEventID(c *gin.Context, idEvent int64) {
	users, err := s.userService.GetDummiesByEventID(c.Request.Context(), idEvent)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении dummy-пользователей: %w", err))
		return
	}

	apiUsers := make([]api.UserProfileDTO, 0, len(users))
	for i := range users {
		apiUsers = append(apiUsers, convertUserToProfileAPI(&users[i]))
	}

	c.JSON(http.StatusOK, api.UserListResponse{Users: &apiUsers})
}

// CreateDummyUser создает dummy-пользователя
func (s *ServerHandler) CreateDummyUser(c *gin.Context, idEvent int64) {
	var apiRequest api.DummyUserRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "некорректные данные запроса",
		})
		return
	}

	user, err := s.userService.CreateDummyUser(c.Request.Context(), apiRequest.Nickname, idEvent)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при создании dummy-пользователя: %w", err))
		return
	}

	c.JSON(http.StatusCreated, convertUserToProfileAPI(user))
}

// AddUsersToEvent добавляет пользователей в мероприятие
func (s *ServerHandler) AddUsersToEvent(c *gin.Context, idEvent int64) {
	var apiRequest api.AddUsersRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "некорректные данные запроса",
		})
		return
	}

	if len(apiRequest.UserIds) == 0 {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "список пользователей не может быть пустым",
		})
		return
	}

	err := s.userService.AddUsersToEvent(c.Request.Context(), apiRequest.UserIds, idEvent)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при добавлении пользователей: %w", err))
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{Success: true})
}

// RemoveUserFromEvent удаляет пользователя из мероприятия
func (s *ServerHandler) RemoveUserFromEvent(c *gin.Context, idEvent int64, idUser int64) {
	err := s.userService.RemoveUserFromEvent(c.Request.Context(), idUser, idEvent)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при удалении пользователя: %w", err))
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{Success: true})
}

// SyncUsers синхронизирует пользователей с ff-id сервисом
func (s *ServerHandler) SyncUsers(c *gin.Context) {
	var apiRequest api.SyncUsersRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "некорректные данные запроса",
		})
		return
	}

	if len(apiRequest.UserIds) == 0 {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Error: "список пользователей не может быть пустым",
		})
		return
	}

	err := s.userService.BatchSyncUsersWithIDService(c.Request.Context(), apiRequest.UserIds)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при синхронизации пользователей: %w", err))
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{Success: true})
}

// Helper functions для конвертации типов

func convertUserToProfileAPI(user *models.User) api.UserProfileDTO {
	return api.UserProfileDTO{
		InternalId: &user.ID,
		UserId:     user.UserID,
		Nickname:   &user.NicknameCashed,
		Name:       &user.NameCashed,
		Photo:      &user.PhotoUUIDCashed,
	}
}

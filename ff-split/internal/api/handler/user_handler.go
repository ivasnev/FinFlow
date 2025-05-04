package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// UserHandler обработчик для работы с пользователями
type UserHandler struct {
	service service.UserServiceInterface
}

// NewUserHandler создает новый обработчик для работы с пользователями
func NewUserHandler(service service.UserServiceInterface) *UserHandler {
	return &UserHandler{service: service}
}

// GetUserByID возвращает пользователя по ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id_user")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID пользователя"})
		return
	}

	byExternal := c.Query("by_external")
	var user *models.User
	if byExternal == "true" {
		user, err = h.service.GetUserByExternalUserID(c.Request.Context(), id)
	} else {
		user, err = h.service.GetUserByInternalUserID(c.Request.Context(), id)
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapUserToResponse(user))
}

// GetUsersByIDs возвращает пользователей по списку ID
func (h *UserHandler) GetUsersByIDs(c *gin.Context) {
	var request struct {
		UserIDs []int64 `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	byExternal := c.Query("by_external")
	var users []models.User
	var err error
	if byExternal == "true" {
		users, err = h.service.GetUsersByExternalUserIDs(c.Request.Context(), request.UserIDs)
	} else {
		users, err = h.service.GetUsersByInternalUserIDs(c.Request.Context(), request.UserIDs)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.UserResponse, len(users))
	for i, user := range users {
		response[i] = mapUserToResponse(&user)
	}

	c.JSON(http.StatusOK, response)
}

// GetUsersByEventID возвращает пользователей мероприятия
func (h *UserHandler) GetUsersByEventID(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID мероприятия"})
		return
	}

	users, err := h.service.GetUsersByEventID(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.UserResponse, len(users))
	for i, user := range users {
		response[i] = mapUserToResponse(&user)
	}

	c.JSON(http.StatusOK, response)
}

// GetDummiesByEventID возвращает фиктивных пользователей мероприятия
func (h *UserHandler) GetDummiesByEventID(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID мероприятия"})
		return
	}

	users, err := h.service.GetDummiesByEventID(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.UserResponse, len(users))
	for i, user := range users {
		response[i] = mapUserToResponse(&user)
	}

	c.JSON(http.StatusOK, response)
}

// SyncUsers синхронизирует пользователей с ID-сервисом
func (h *UserHandler) SyncUsers(c *gin.Context) {
	var request struct {
		UserIDs []int64 `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.BatchSyncUsersWithIDService(c.Request.Context(), request.UserIDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// CreateDummyUser создает фиктивного пользователя
func (h *UserHandler) CreateDummyUser(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID мероприятия"})
		return
	}

	var request struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.CreateDummyUser(c.Request.Context(), request.Name, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, mapUserToResponse(user))
}

// BatchCreateDummyUsers создает несколько фиктивных пользователей
func (h *UserHandler) BatchCreateDummyUsers(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID мероприятия"})
		return
	}

	var request struct {
		Names []string `json:"names" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users, err := h.service.BatchCreateDummyUsers(c.Request.Context(), request.Names, eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.UserResponse, len(users))
	for i, user := range users {
		response[i] = mapUserToResponse(user)
	}

	c.JSON(http.StatusCreated, response)
}

// UpdateUser обновляет пользователя
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id_user")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID пользователя"})
		return
	}

	var request struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Сначала получаем пользователя
	user, err := h.service.GetUserByInternalUserID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Обновляем данные
	user.NameCashed = request.Name

	// Сохраняем изменения
	updatedUser, err := h.service.UpdateUser(c.Request.Context(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapUserToResponse(updatedUser))
}

// DeleteUser удаляет пользователя
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id_user")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID пользователя"})
		return
	}

	if err := h.service.DeleteUser(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddUsersToEvent добавляет пользователей в мероприятие
func (h *UserHandler) AddUsersToEvent(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID мероприятия"})
		return
	}

	var request struct {
		UserIDs []int64 `json:"user_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddUsersToEvent(c.Request.Context(), request.UserIDs, eventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// RemoveUserFromEvent удаляет пользователя из мероприятия
func (h *UserHandler) RemoveUserFromEvent(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID мероприятия"})
		return
	}

	idStr := c.Param("id_user")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID пользователя"})
		return
	}

	if err := h.service.RemoveUserFromEvent(c.Request.Context(), userID, eventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// Вспомогательные функции

// mapUserToResponse преобразует модель пользователя в DTO ответа
func mapUserToResponse(user *models.User) dto.UserResponse {
	response := dto.UserResponse{
		ID:      user.ID,
		Name:    user.NameCashed,
		IsDummy: user.IsDummy,
	}

	// Дополняем информацию о профиле для не-dummy пользователей
	if !user.IsDummy && user.UserID != nil {
		response.Profile = &dto.UserProfileDTO{
			UserID:   *user.UserID,
			Nickname: user.NicknameCashed,
			Name:     user.NameCashed,
			Photo:    user.PhotoUUIDCashed,
		}
	}

	return response
}

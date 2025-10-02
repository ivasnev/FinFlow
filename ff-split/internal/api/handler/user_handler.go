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
// @Summary Получить пользователя по ID
// @Description Возвращает информацию о пользователе по его ID
// @Tags пользователи
// @Accept json
// @Produce json
// @Param id_user path int true "ID пользователя"
// @Param by_external query bool false "Искать по внешнему ID" default(false)
// @Success 200 {object} dto.UserResponse "Информация о пользователе"
// @Failure 400 {object} map[string]string "Неверный формат ID пользователя"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Router /api/v1/user/{id_user} [get]
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
// @Summary Получить пользователей по списку ID
// @Description Возвращает информацию о пользователях по списку ID
// @Tags пользователи
// @Accept json
// @Produce json
// @Param by_external query bool false "Искать по внешним ID" default(false)
// @Param request body object true "Список ID пользователей"
// @Success 200 {array} dto.UserResponse "Список пользователей"
// @Failure 400 {object} map[string]string "Неверный формат запроса"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/user/list [post]
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
// @Summary Получить пользователей мероприятия
// @Description Возвращает список всех пользователей, связанных с мероприятием
// @Tags пользователи
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Success 200 {array} dto.UserResponse "Список пользователей"
// @Failure 400 {object} map[string]string "Неверный формат ID мероприятия"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/user [get]
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
// @Summary Получить фиктивных пользователей мероприятия
// @Description Возвращает список всех фиктивных (dummy) пользователей, связанных с мероприятием
// @Tags пользователи
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Success 200 {array} dto.UserResponse "Список фиктивных пользователей"
// @Failure 400 {object} map[string]string "Неверный формат ID мероприятия"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/user/dummies [get]
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
// @Summary Синхронизировать пользователей с ID-сервисом
// @Description Обновляет локальные данные пользователей из ID-сервиса
// @Tags пользователи
// @Accept json
// @Produce json
// @Param request body object true "Список ID пользователей для синхронизации"
// @Success 200 "Пользователи успешно синхронизированы"
// @Failure 400 {object} map[string]string "Неверный формат запроса"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/user/sync [post]
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
// @Summary Создать фиктивного пользователя
// @Description Создает нового фиктивного (dummy) пользователя в рамках мероприятия
// @Tags пользователи
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param request body object true "Данные фиктивного пользователя"
// @Success 201 {object} dto.UserResponse "Созданный фиктивный пользователь"
// @Failure 400 {object} map[string]string "Неверный формат запроса"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/user/dummy [post]
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
// @Summary Создать несколько фиктивных пользователей
// @Description Создает несколько фиктивных (dummy) пользователей в рамках мероприятия
// @Tags пользователи
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param request body object true "Список имен фиктивных пользователей"
// @Success 201 {array} dto.UserResponse "Список созданных фиктивных пользователей"
// @Failure 400 {object} map[string]string "Неверный формат запроса"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/user/dummies [post]
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
// @Summary Обновить пользователя
// @Description Обновляет информацию о пользователе
// @Tags пользователи
// @Accept json
// @Produce json
// @Param id_user path int true "ID пользователя"
// @Param request body object true "Данные для обновления"
// @Success 200 {object} dto.UserResponse "Обновленный пользователь"
// @Failure 400 {object} map[string]string "Неверный формат запроса"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/user/{id_user} [put]
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
// @Summary Удалить пользователя
// @Description Удаляет пользователя по ID
// @Tags пользователи
// @Accept json
// @Produce json
// @Param id_user path int true "ID пользователя"
// @Success 204 "Пользователь успешно удален"
// @Failure 400 {object} map[string]string "Неверный формат ID пользователя"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/user/{id_user} [delete]
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
// @Summary Добавить пользователей в мероприятие
// @Description Добавляет список пользователей в мероприятие
// @Tags пользователи
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param request body object true "Список ID пользователей"
// @Success 200 "Пользователи успешно добавлены"
// @Failure 400 {object} map[string]string "Неверный формат запроса"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/user [post]
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
// @Summary Удалить пользователя из мероприятия
// @Description Удаляет пользователя из мероприятия
// @Tags пользователи
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id_user path int true "ID пользователя"
// @Success 204 "Пользователь успешно удален из мероприятия"
// @Failure 400 {object} map[string]string "Неверный формат ID"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/user/{id_user} [delete]
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

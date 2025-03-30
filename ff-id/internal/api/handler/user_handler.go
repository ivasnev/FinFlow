package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/dto"
	"github.com/ivasnev/FinFlow/ff-id/interfaces"
)

// UserHandler обрабатывает запросы, связанные с пользователями
type UserHandler struct {
	userService interfaces.UserService
}

// NewUserHandler создает новый UserHandler
func NewUserHandler(userService interfaces.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// GetUserByNickname обрабатывает запрос на получение информации о пользователе по никнейму
// @Summary Get user by nickname
// @Description Get user information by nickname
// @Tags users
// @Produce json
// @Param nickname path string true "User nickname"
// @Success 200 {object} dto.UserDTO
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /users/{nickname} [get]
func (h *UserHandler) GetUserByNickname(c *gin.Context) {
	nickname := c.Param("nickname")

	user, err := h.userService.GetUserByNickname(c.Request.Context(), nickname)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser обрабатывает запрос на обновление профиля пользователя
// @Summary Update user profile
// @Description Update current user's profile
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.UpdateUserRequest true "User update data"
// @Success 200 {object} dto.UserDTO
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /users/me [patch]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID пользователя из контекста, установленного middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Преобразуем ID в int64, если необходимо
	var userIDInt64 int64
	switch v := userID.(type) {
	case int64:
		userIDInt64 = v
	case float64:
		userIDInt64 = int64(v)
	case string:
		var err error
		userIDInt64, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID format"})
			return
		}
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), userIDInt64, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

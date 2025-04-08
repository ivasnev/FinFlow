package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
)

// UserHandler обрабатывает запросы, связанные с пользователями
type UserHandler struct {
	userService service.UserServiceInterface
}

// NewUserHandler создает новый UserHandler
func NewUserHandler(userService service.UserServiceInterface) *UserHandler {
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
// @Description Update user's profile by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body dto.UpdateUserRequest true "User update data"
// @Success 200 {object} dto.UserDTO
// @Failure 400 {object} gin.H
// @Failure 404 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /users/{id} [patch]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Получаем ID пользователя из URL
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

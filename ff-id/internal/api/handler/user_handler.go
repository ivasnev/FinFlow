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

// GetUsersByIds обрабатывает запрос на получение информации о пользователях по их ID
// @Summary Get users by IDs
// @Description Get user information by their IDs
// @Tags users
// @Produce json
// @Param ids query []int64 true "User IDs"
// @Success 200 {object} []dto.UserDTO
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /internal/users [get]
func (h *UserHandler) GetUsersByIds(c *gin.Context) {
	var err error
	ids := c.QueryArray("user_id")
	idsInt64 := make([]int64, len(ids))
	for i, id := range ids {
		idsInt64[i], err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
			return
		}
	}

	users, err := h.userService.GetUsersByIds(c.Request.Context(), idsInt64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
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
// @Router /users/me [patch]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	// Получаем ID пользователя из URL
	userIDStr, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found in context"})
	}
	userID, canParse := userIDStr.(int64)
	if !canParse {
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

// RegisterUser обрабатывает запрос на регистрацию пользователя от клиента с токеном авторизации
// @Summary Регистрация нового пользователя
// @Description Регистрирует нового пользователя с использованием токена авторизации
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.RegisterUserRequest true "Данные для регистрации пользователя"
// @Success 201 {object} dto.RegisterUserResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /users/register [post]
func (h *UserHandler) RegisterUser(c *gin.Context) {
	// Получаем ID пользователя из контекста (установлен middleware)
	userIDStr, exist := c.Get("user_id")
	if !exist {
		c.JSON(http.StatusBadRequest, gin.H{"error": "отсутствует ID пользователя в контексте"})
		return
	}
	userID, ok := userIDStr.(int64)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный формат ID пользователя"})
		return
	}

	// Парсим данные из запроса
	var req *dto.RegisterUserRequest
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Регистрируем пользователя
	user, err := h.userService.RegisterUser(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusCreated, user)
}

// RegisterUserFromService обрабатывает запрос на регистрацию пользователя от другого сервиса
// @Summary Регистрация нового пользователя от сервиса
// @Description Регистрирует нового пользователя от имени другого сервиса (TVM)
// @Tags users
// @Accept json
// @Produce json
// @Param request body dto.ServiceRegisterUserRequest true "Данные для регистрации пользователя"
// @Success 201 {object} dto.RegisterUserResponse
// @Failure 400 {object} gin.H
// @Failure 500 {object} gin.H
// @Router /internal/users/register [post]
func (h *UserHandler) RegisterUserFromService(c *gin.Context) {

	// Парсим данные из запроса
	var req dto.ServiceRegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req_dto := &dto.RegisterUserRequest{
		Email:    req.Email,
		Nickname: req.Nickname,
	}

	// Регистрируем пользователя
	user, err := h.userService.RegisterUser(c.Request.Context(), req.UserID, req_dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusCreated, user)
}

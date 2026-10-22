package handler

import (
	"encoding/base64"
	"errors"
	"github.com/ivasnev/FinFlow/ff-auth/internal/account"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/internal/adapters/ffid"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service"
	"github.com/ivasnev/FinFlow/ff-auth/internal/service/password"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/api"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ServerHandler реализует интерфейс api.ServerInterface
type ServerHandler struct {
	Account             *account.Manager
	authService         service.Auth
	userService         service.User
	sessionService      service.Session
	loginHistoryService service.LoginHistory
	tokenManager        service.TokenManager
}

// NewServerHandler создает новый ServerHandler
func NewServerHandler(
	authService service.Auth,
	userService service.User,
	sessionService service.Session,
	loginHistoryService service.LoginHistory,
	tokenManager service.TokenManager,
) *ServerHandler {
	return &ServerHandler{
		authService:         authService,
		userService:         userService,
		sessionService:      sessionService,
		loginHistoryService: loginHistoryService,
		tokenManager:        tokenManager,
	}
}

// Login обрабатывает запрос на вход в систему
func (h *ServerHandler) Login(c *gin.Context) {
	if h.Account != nil {
		h.Account.Login(c)
		return
	}
	var req api.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в сервисные типы
	loginParams := service.LoginParams{
		Login:     req.Login,
		Password:  req.Password,
		UserAgent: c.GetHeader("User-Agent"),
		IpAddress: c.ClientIP(),
	}

	response, err := h.authService.Login(c.Request.Context(), loginParams)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, ffid.ErrUnavailable) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем ответ в API типы
	apiResponse := api.AuthResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresAt:    response.ExpiresAt,
		User: api.ShortUserDTO{
			Id:       response.User.Id,
			Email:    openapi_types.Email(response.User.Email),
			Nickname: response.User.Nickname,
			Roles:    response.User.Roles,
		},
	}

	c.JSON(http.StatusOK, apiResponse)
}

// Logout обрабатывает запрос на выход из системы
func (h *ServerHandler) Logout(c *gin.Context) {
	var req api.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	err := h.authService.Logout(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}

// GetPublicKey возвращает публичный ключ
func (h *ServerHandler) GetPublicKey(c *gin.Context) {
	publicKey := h.tokenManager.GetPublicKey()
	encodedKey := base64.StdEncoding.EncodeToString(publicKey)
	c.String(http.StatusOK, encodedKey)
}

// RefreshToken обрабатывает запрос на обновление access-токена
func (h *ServerHandler) RefreshToken(c *gin.Context) {
	if h.Account != nil {
		h.Account.Refresh(c)
		return
	}
	var req api.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	response, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, ffid.ErrUnavailable) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем ответ в API типы
	apiResponse := api.AuthResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresAt:    response.ExpiresAt,
		User: api.ShortUserDTO{
			Id:       response.User.Id,
			Email:    openapi_types.Email(response.User.Email),
			Nickname: response.User.Nickname,
			Roles:    response.User.Roles,
		},
	}

	c.JSON(http.StatusOK, apiResponse)
}

// Register обрабатывает запрос на регистрацию нового пользователя
func (h *ServerHandler) Register(c *gin.Context) {
	if h.Account != nil {
		h.Account.Register(c)
		return
	}
	var req api.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в сервисные типы
	registerParams := service.RegisterParams{
		Email:    string(req.Email),
		Phone:    req.Phone,
		Password: req.Password,
		Nickname: req.Nickname,
		Name:     req.Name,
	}

	response, err := h.authService.Register(c.Request.Context(), registerParams)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ffid.ErrConflict) {
			status = http.StatusConflict
		}
		if errors.Is(err, ffid.ErrUnavailable) {
			status = http.StatusServiceUnavailable
		}
		if errors.Is(err, password.ErrInvalid) {
			status = http.StatusBadRequest
		}
		c.JSON(status, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем ответ в API типы
	apiResponse := api.AuthResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
		ExpiresAt:    response.ExpiresAt,
		User: api.ShortUserDTO{
			Id:       response.User.Id,
			Email:    openapi_types.Email(response.User.Email),
			Nickname: response.User.Nickname,
			Roles:    response.User.Roles,
		},
	}

	c.JSON(http.StatusCreated, apiResponse)
}

// GetLoginHistory обрабатывает запрос на получение истории входов
func (h *ServerHandler) GetLoginHistory(c *gin.Context, params api.GetLoginHistoryParams) {
	// Получаем данные пользователя из контекста
	userData, exists := auth.GetUserData(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Устанавливаем значения по умолчанию для пагинации
	limit := 10
	offset := 0

	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	history, err := h.loginHistoryService.GetUserLoginHistory(c.Request.Context(), userData.UserID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в API типы
	apiHistory := make([]api.LoginHistoryDTO, len(history))
	for i, item := range history {
		apiHistory[i] = api.LoginHistoryDTO{
			Id:        item.Id,
			IpAddress: item.IpAddress,
			UserAgent: item.UserAgent,
			CreatedAt: item.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, apiHistory)
}

// GetUserSessions обрабатывает запрос на получение активных сессий пользователя
func (h *ServerHandler) GetUserSessions(c *gin.Context) {
	// Получаем данные пользователя из контекста
	userData, exists := auth.GetUserData(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "unauthorized"})
		return
	}

	sessions, err := h.sessionService.GetUserSessions(c.Request.Context(), userData.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем в API типы
	apiSessions := make([]api.SessionDTO, len(sessions))
	for i, session := range sessions {
		apiSessions[i] = api.SessionDTO{
			Id:        openapi_types.UUID(session.Id),
			IpAddress: session.IpAddress,
			CreatedAt: session.CreatedAt,
			ExpiresAt: session.ExpiresAt,
		}
	}

	c.JSON(http.StatusOK, apiSessions)
}

// TerminateSession обрабатывает запрос на завершение сессии
func (h *ServerHandler) TerminateSession(c *gin.Context, id openapi_types.UUID) {
	// Получаем данные пользователя из контекста
	userData, exists := auth.GetUserData(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "unauthorized"})
		return
	}

	err := h.sessionService.TerminateSession(c.Request.Context(), id, userData.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session terminated successfully"})
}

// UpdateUser обрабатывает запрос на обновление профиля пользователя
func (h *ServerHandler) UpdateUser(c *gin.Context) {
	var req api.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Получаем данные пользователя из контекста
	userData, exists := auth.GetUserData(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, api.ErrorResponse{Error: "unauthorized"})
		return
	}

	// Конвертируем в сервисные типы
	var email *string
	if req.Email != nil {
		emailStr := string(*req.Email)
		email = &emailStr
	}

	if h.Account != nil && (req.Email != nil || req.Password != nil) {
		c.JSON(400, gin.H{"error": "Используйте отдельную подтверждаемую смену email или пароля в настройках безопасности."})
		return
	}
	updateData := service.UserUpdateData{
		Email:    email,
		Nickname: req.Nickname,
		Password: req.Password,
	}

	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), userData.UserID, updateData)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ffid.ErrConflict) {
			status = http.StatusConflict
		}
		if errors.Is(err, ffid.ErrUnavailable) {
			status = http.StatusServiceUnavailable
		}
		if errors.Is(err, password.ErrInvalid) {
			status = http.StatusBadRequest
		}
		c.JSON(status, api.ErrorResponse{Error: err.Error()})
		return
	}

	// Конвертируем ответ в API типы
	apiUser := api.UserDTO{
		Id:        updatedUser.Id,
		Email:     openapi_types.Email(updatedUser.Email),
		Nickname:  updatedUser.Nickname,
		Roles:     updatedUser.Roles,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	}

	c.JSON(http.StatusOK, apiUser)
}

// GetUserByNickname обрабатывает запрос на получение пользователя по nickname
func (h *ServerHandler) GetUserByNickname(c *gin.Context, nickname string) {
	user, err := h.userService.GetUserByNickname(c.Request.Context(), nickname)
	if err != nil {
		status := http.StatusNotFound
		if errors.Is(err, ffid.ErrUnavailable) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, api.ErrorResponse{Error: "user not found or unavailable"})
		return
	}

	// Конвертируем ответ в API типы
	apiUser := api.UserDTO{
		Id:        user.Id,
		Email:     openapi_types.Email(user.Email),
		Nickname:  user.Nickname,
		Roles:     user.Roles,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, apiUser)
}

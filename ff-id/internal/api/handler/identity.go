package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-id/internal/repository"
	"github.com/ivasnev/FinFlow/ff-id/internal/service"
	"github.com/ivasnev/FinFlow/ff-id/pkg/api"
)

// LookupIdentity возвращает профиль из единственного источника актуальных email и nickname.
func (h *ServerHandler) LookupIdentity(c *gin.Context, params api.LookupIdentityParams) {
	if (params.Login == nil) == (params.UserId == nil) || params.UserId != nil && *params.UserId <= 0 {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "требуется login либо положительный user_id"})
		return
	}
	var user *service.UserDTO
	var err error
	kind := "login"
	if params.Kind != nil {
		kind = string(*params.Kind)
	}
	if kind != "login" && kind != "email" && kind != "nickname" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "неизвестный тип поиска"})
		return
	}
	switch {
	case params.UserId != nil:
		user, err = h.userService.GetUserByID(c.Request.Context(), *params.UserId)
	case kind == "email" || kind == "login" && strings.Contains(*params.Login, "@"):
		user, err = h.userService.GetUserByEmail(c.Request.Context(), *params.Login)
	default:
		user, err = h.userService.GetUserByNickname(c.Request.Context(), *params.Login)
	}

	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, repository.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, api.ErrorResponse{Error: "пользователь не найден или недоступен"})
		return
	}
	c.JSON(http.StatusOK, convertToAPIUserDTO(user))
}

// UpdateIdentity изменяет email и nickname в ff-id от имени доверенного сервиса.
func (h *ServerHandler) UpdateIdentity(c *gin.Context, userID int64) {
	var req api.UpdateIdentityJSONRequestBody
	if userID <= 0 || c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "некорректные данные"})
		return
	}
	var email *string
	if req.Email != nil {
		value := string(*req.Email)
		email = &value
	}
	user, err := h.userService.UpdateUser(c.Request.Context(), userID, service.UpdateUserRequest{Email: email, Nickname: req.Nickname})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, service.ErrIdentityConflict) {
			status = http.StatusConflict
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, api.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, convertToAPIUserDTO(user))
}

package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
)

// invitationResponse не раскрывает хеш и секрет в списке ссылок.
func invitationResponse(row models.Invitation) api.InvitationResponse {
	return api.InvitationResponse{Id: row.ID, EventId: row.EventID, CreatedBy: row.CreatedBy, CreatedAt: row.CreatedAt, ExpiresAt: row.ExpiresAt}
}

// ListInvitations отдаёт собственные активные ссылки после проверки участия.
func (s *ServerHandler) ListInvitations(c *gin.Context, eventID int64) {
	rows, err := s.invitationService.List(c.Request.Context(), eventID)
	if err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	result := api.InvitationListResponse{Invitations: make([]api.InvitationResponse, 0, len(rows))}
	for _, row := range rows {
		result.Invitations = append(result.Invitations, invitationResponse(row))
	}
	c.JSON(http.StatusOK, result)
}

// CreateInvitation выдаёт секрет только при создании и запрещает кэширование ответа.
func (s *ServerHandler) CreateInvitation(c *gin.Context, eventID int64) {
	c.Header("Cache-Control", "no-store")
	row, token, err := s.invitationService.Create(c.Request.Context(), eventID)
	if err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusCreated, api.InvitationCreatedResponse{Invitation: invitationResponse(*row), Token: token})
}

// RevokeInvitation отзывает ссылку без удаления её истории.
func (s *ServerHandler) RevokeInvitation(c *gin.Context, eventID, id int64) {
	if err := s.invitationService.Revoke(c.Request.Context(), eventID, id); err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api.SuccessResponse{Success: true})
}

// invitationToken читает секрет из ограниченного тела, не помещая его в URL и журналы запросов.
func invitationToken(c *gin.Context) (string, bool) {
	c.Header("Cache-Control", "no-store")
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var request api.InvitationTokenRequest
	if err := decoder.Decode(&request); err != nil {
		errors.HTTPErrorHandler(c, errors.NewValidationError("invitation", "некорректное приглашение"))
		return "", false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		errors.HTTPErrorHandler(c, errors.NewValidationError("invitation", "ожидается одно приглашение"))
		return "", false
	}
	return request.Token, true
}

// PreviewInvitation показывает название без изменения состава участников.
func (s *ServerHandler) PreviewInvitation(c *gin.Context) {
	token, ok := invitationToken(c)
	if !ok {
		return
	}
	row, err := s.invitationService.Preview(c.Request.Context(), token)
	if err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api.InvitationPreviewResponse{EventId: row.EventID, EventName: row.EventName, ExpiresAt: row.ExpiresAt, IsMember: row.IsMember})
}

// AcceptInvitation добавляет только пользователя из подтверждённого контекста авторизации.
func (s *ServerHandler) AcceptInvitation(c *gin.Context) {
	token, ok := invitationToken(c)
	if !ok {
		return
	}
	id, err := s.invitationService.Accept(c.Request.Context(), token)
	if err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api.InvitationAcceptedResponse{EventId: id})
}

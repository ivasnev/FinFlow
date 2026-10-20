package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
	"net/http"
)

// GetArchivedEvents показывает только события, в которых пользователь сохранил участие.
func (s *ServerHandler) GetArchivedEvents(c *gin.Context) {
	rows, err := s.archiveService.List(c.Request.Context())
	if err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	events := make([]api.EventResponse, 0, len(rows))
	for _, row := range rows {
		row := row
		events = append(events, api.EventResponse{Id: &row.ID, Name: &row.Name, Description: &row.Description, CategoryId: row.CategoryID, PhotoId: &row.ImageID, DeletedAt: row.DeletedAt, PurgeAt: row.PurgeAt})
	}
	c.JSON(http.StatusOK, api.EventListResponse{Events: &events})
}

// RestoreEvent возвращает доступ к сохранённым данным и не продлевает активные приглашения.
func (s *ServerHandler) RestoreEvent(c *gin.Context, idEvent int64) {
	if err := s.archiveService.Restore(c.Request.Context(), idEvent); err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api.SuccessResponse{Success: true})
}

package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
	"net/http"
)

// ListHistory отдаёт историю только после стандартной проверки участия в событии.
func (s *ServerHandler) ListHistory(c *gin.Context, eventID int64, params api.ListHistoryParams) {
	var before int64
	if params.Before != nil {
		before = *params.Before
		if before <= 0 {
			errors.HTTPErrorHandler(c, errors.NewValidationError("before", "некорректная страница"))
			return
		}
	}
	rows, err := s.historyService.List(c.Request.Context(), eventID, before)
	if err != nil {
		errors.HTTPErrorHandler(c, err)
		return
	}
	result := api.HistoryListResponse{Entries: make([]api.HistoryEntry, 0, len(rows))}
	for _, r := range rows {
		result.Entries = append(result.Entries, api.HistoryEntry{Id: r.ID, EventId: r.EventID, ActorId: r.ActorID, ActorName: r.ActorName, Kind: r.Kind, EntityId: r.EntityID, Before: r.BeforeValues, After: r.AfterValues, CreatedAt: r.CreatedAt})
	}
	if len(rows) == 50 {
		id := rows[len(rows)-1].ID
		result.NextBefore = &id
	}
	c.JSON(http.StatusOK, result)
}

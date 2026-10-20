package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	domainerrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
)

// settlementResponse переводит журнал в публичный контракт без изменения денежных единиц.
func settlementResponse(row models.Settlement) api.SettlementResponse {
	return api.SettlementResponse{Id: row.ID, EventId: row.EventID, FromUserId: row.FromUserID, ToUserId: row.ToUserID, CreatedBy: row.CreatedBy, AmountCents: row.AmountCents, RequestId: uuid.MustParse(row.RequestID), CreatedAt: row.CreatedAt, CanceledAt: row.CanceledAt}
}

// ListSettlements возвращает ограниченную страницу журнала после проверки участия middleware.
func (s *ServerHandler) ListSettlements(c *gin.Context, idEvent int64, params api.ListSettlementsParams) {
	var before int64
	if params.Before != nil {
		before = *params.Before
		if before <= 0 {
			domainerrors.HTTPErrorHandler(c, domainerrors.NewValidationError("before", "некорректная страница"))
			return
		}
	}
	rows, err := s.settlementService.List(c.Request.Context(), idEvent, before)
	if err != nil {
		domainerrors.HTTPErrorHandler(c, err)
		return
	}
	result := api.SettlementListResponse{Settlements: make([]api.SettlementResponse, 0, len(rows))}
	for _, row := range rows {
		result.Settlements = append(result.Settlements, settlementResponse(row))
	}
	if len(rows) == 50 {
		id := rows[len(rows)-1].ID
		result.NextBefore = &id
	}
	c.JSON(http.StatusOK, result)
}

// ReceiveSettlement определяет получателя из доверенного токена, а не из тела запроса.
func (s *ServerHandler) ReceiveSettlement(c *gin.Context, idEvent int64) {
	actor, ok := auth.GetUserData(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 2048)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var request api.SettlementRequest
	if err := decoder.Decode(&request); err != nil {
		domainerrors.HTTPErrorHandler(c, domainerrors.NewValidationError("settlement", "некорректные данные возврата"))
		return
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		domainerrors.HTTPErrorHandler(c, domainerrors.NewValidationError("settlement", "ожидается один запрос"))
		return
	}
	row, err := s.settlementService.Receive(c.Request.Context(), idEvent, actor.UserID, service.SettlementRequest{FromUserID: request.FromUserId, AmountCents: request.AmountCents, RequestID: request.RequestId.String()})
	if err != nil {
		domainerrors.HTTPErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, settlementResponse(*row))
}

// CancelSettlement сохраняет запись с отметкой отмены и проверяет автора внутри SQL-транзакции.
func (s *ServerHandler) CancelSettlement(c *gin.Context, idEvent, idSettlement int64) {
	actor, ok := auth.GetUserData(c)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if err := s.settlementService.Cancel(c.Request.Context(), idEvent, actor.UserID, idSettlement); err != nil {
		domainerrors.HTTPErrorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api.SuccessResponse{Success: true})
}

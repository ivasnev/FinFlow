package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
	"github.com/stretchr/testify/require"
)

// TestEventBalanceCentsHTTP проверяет точные балансы списка и карточки через PostgreSQL.
func TestEventBalanceCentsHTTP(t *testing.T) {
	database, c := setupOptimizationTest(t)
	require.NoError(t, database.DB.Exec(`UPDATE users SET user_id=100+id,is_dummy=false;
 INSERT INTO user_event(user_id,event_id) VALUES(1,2),(2,2),(3,2);`).Error)
	_, err := c.TransactionService.CreateTransaction(context.Background(), 1, optimizationRequest(120.04))
	require.NoError(t, err)
	h := handler.NewServerHandler(c.EventService, c.UserService, c.TransactionService, c.ActivityService, c.TaskService, c.CategoryService, c.IconService, nil)
	api.RegisterHandlersWithOptions(c.Router, h, api.GinServerOptions{Middlewares: []api.MiddlewareFunc{
		func(ctx *gin.Context) {
			accountID, err := strconv.ParseInt(ctx.GetHeader("X-Test-Account"), 10, 64)
			require.NoError(t, err)
			ctx.Set(string(auth.UserContextKey()), auth.UserData{UserID: accountID, Roles: []string{"user"}})
		},
	}})
	for _, tc := range []struct {
		account int
		want    float64
	}{{101, 60.02}, {102, -60.02}, {103, 0}} {
		// Сценарий: получатель, должник и участник без долгов видят баланс без потери копеек.
		t.Run(fmt.Sprint(tc.account), func(t *testing.T) {
			for _, path := range []string{"/api/v1/event", "/api/v1/event/1", "/api/v1/event/2"} {
				req := httptest.NewRequest(http.MethodGet, path, nil)
				req.Header.Set("X-Test-Account", strconv.Itoa(tc.account))
				response := httptest.NewRecorder()
				c.Router.ServeHTTP(response, req)
				require.Equal(t, 200, response.Code, response.Body.String())
				var events []api.EventResponse
				if path == "/api/v1/event" {
					var list api.EventListResponse
					require.NoError(t, json.Unmarshal(response.Body.Bytes(), &list))
					require.NotNil(t, list.Events)
					events = *list.Events
					require.Len(t, events, 2)
				} else {
					var event api.EventResponse
					require.NoError(t, json.Unmarshal(response.Body.Bytes(), &event))
					events = []api.EventResponse{event}
				}
				for _, event := range events {
					require.NotNil(t, event.Id)
					require.NotNil(t, event.Balance)
					want := tc.want
					if *event.Id == 2 {
						want = 0
					}
					require.Equal(t, want, *event.Balance, "path=%s event=%d", path, *event.Id)
				}
			}
		})
	}
}

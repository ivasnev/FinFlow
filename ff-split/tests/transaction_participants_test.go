package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
)

// TestTransactionParticipants проверяет чужих участников при создании и изменении расхода через HTTP и PostgreSQL.
func TestTransactionParticipants(t *testing.T) {
	database := setupTestDB(t)
	t.Cleanup(func() { teardownTestDB(t, database) })
	db := database.DB
	if err := db.Exec(`INSERT INTO users(id,user_id,is_dummy) VALUES(10,100,false),(20,NULL,true),(30,300,false);
		INSERT INTO events(id,name) VALUES(1,'own'),(2,'foreign');
		INSERT INTO user_event(user_id,event_id) VALUES(10,1),(20,1),(30,2);`).Error; err != nil {
		t.Fatal(err)
	}
	c, err := createTestContainer(t, &config.Config{}, gin.New(), db, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	h := handler.NewServerHandler(c.EventService, c.UserService, c.TransactionService, c.ActivityService, c.TaskService, c.CategoryService, c.IconService)
	api.RegisterHandlers(c.Router, h)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		c.Router.ServeHTTP(res, req)
		return res
	}
	// Сценарий: плательщик является участником мероприятия, но оплачивает только другого участника.
	if res := request("POST", "/api/v1/event/1/transaction", `{"name":"original","amount":100,"from_user":10,"users":[20],"type":"units"}`); res.Code != 201 {
		t.Fatalf("valid create: %d %s", res.Code, res.Body)
	}
	for _, kind := range []string{"percent", "amount", "units"} {
		for _, tc := range []struct {
			name, users, portion string
			payer                int
		}{
			{"foreign portion", `[10,20]`, `{"10":50,"30":50}`, 10},
			{"foreign consumer", `[10,30]`, `{"10":50,"30":50}`, 10},
			{"foreign payer", `[10,20]`, `{"10":50,"20":50}`, 30},
			{"external id", `[100,20]`, `{"100":50,"20":50}`, 10},
			{"malformed key", `[10,20]`, `{"10":50,"20evil":50}`, 10},
		} {
			t.Run(kind+"/"+tc.name, func(t *testing.T) {
				body := fmt.Sprintf(`{"name":"invalid","amount":100,"from_user":%d,"users":%s,"type":%q,"portion":%s}`, tc.payer, tc.users, kind, tc.portion)
				for _, endpoint := range []struct{ method, path string }{{"POST", "/api/v1/event/1/transaction"}, {"PUT", "/api/v1/event/1/transaction/1"}} {
					if res := request(endpoint.method, endpoint.path, body); res.Code != 400 {
						t.Fatalf("%s: %d %s", endpoint.method, res.Code, res.Body)
					}
				}
			})
		}
	}
	var state struct {
		Name                           string
		Count, Shares, Debts, Revision int64
		Amount                         float64
	}
	if err := db.Raw(`SELECT name,total_paid AS amount,
		(SELECT COUNT(*) FROM transactions) AS count,
		(SELECT COUNT(*) FROM transaction_shares) AS shares,
		(SELECT COUNT(*) FROM debts) AS debts,
		(SELECT debt_revision FROM events WHERE id=1) AS revision FROM transactions WHERE id=1`).Scan(&state).Error; err != nil {
		t.Fatal(err)
	}
	if state.Name != "original" || state.Amount != 100 || state.Count != 1 || state.Shares != 1 || state.Debts != 1 || state.Revision != 1 {
		t.Fatalf("invalid request changed state: %+v", state)
	}
}

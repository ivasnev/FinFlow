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

// TestTransactionDistribution проверяет HTTP-валидацию и сохранение копеек в PostgreSQL.
func TestTransactionDistribution(t *testing.T) {
	database := setupTestDB(t)
	t.Cleanup(func() { teardownTestDB(t, database) })
	db := database.DB
	if err := db.Exec(`INSERT INTO users(id,is_dummy) VALUES(1,true),(2,true),(3,true);
		INSERT INTO events(id,name) VALUES(1,'trip');
		INSERT INTO user_event(user_id,event_id) VALUES(1,1),(2,1),(3,1);`).Error; err != nil {
		t.Fatal(err)
	}
	c, err := createTestContainer(t, &config.Config{}, gin.New(), db, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	h := handler.NewServerHandler(c.EventService, c.UserService, c.TransactionService, c.ActivityService, c.TaskService, c.CategoryService, c.IconService, nil)
	api.RegisterHandlers(c.Router, h)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		c.Router.ServeHTTP(res, req)
		return res
	}
	for _, tc := range []struct{ splitType, portion string }{
		{"percent", `{"1":-50,"2":150}`},
		{"amount", `{"1":-50,"2":150}`},
		{"units", `{"1":0,"2":0,"3":0}`},
	} {
		body := fmt.Sprintf(`{"name":"invalid","amount":100,"from_user":1,"users":[1,2,3],"type":%q,"portion":%s}`, tc.splitType, tc.portion)
		if res := request("POST", "/api/v1/event/1/transaction", body); res.Code != 400 {
			t.Fatalf("%s: %d %s", tc.splitType, res.Code, res.Body)
		}
	}
	var count int64
	if err := db.Table("transactions").Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("invalid requests persisted: %d %v", count, err)
	}
	// Сценарий: три равных веса сохраняют ровно 100, включая остаточную копейку.
	if res := request("POST", "/api/v1/event/1/transaction", `{"name":"valid","amount":100,"from_user":1,"users":[1,2,3],"type":"units"}`); res.Code != 201 {
		t.Fatalf("create: %d %s", res.Code, res.Body)
	}
	var sums struct{ Total, Shares, Debts float64 }
	if err := db.Raw(`SELECT total_paid AS total,
		(SELECT SUM(value) FROM transaction_shares WHERE transaction_id=t.id) AS shares,
		(SELECT SUM(amount) FROM debts WHERE transaction_id=t.id) AS debts FROM transactions t`).Scan(&sums).Error; err != nil {
		t.Fatal(err)
	}
	if sums.Total != 100 || sums.Shares != 100 || sums.Debts != 66.66 {
		t.Fatalf("lost cents: %+v", sums)
	}
	// Сценарий: невалидное обновление возвращает 400 и сохраняет прежнюю сумму.
	if res := request("PUT", "/api/v1/event/1/transaction/1", `{"name":"invalid","amount":100,"from_user":1,"users":[1,2,3],"type":"units","portion":{"-1":0}}`); res.Code != 400 {
		t.Fatalf("update: %d %s", res.Code, res.Body)
	}
	var total float64
	if err := db.Raw("SELECT total_paid FROM transactions WHERE id=1").Scan(&total).Error; err != nil || total != 100 {
		t.Fatalf("update changed expense: %v %v", total, err)
	}
}

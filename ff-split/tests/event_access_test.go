package tests

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/config"
)

// TestEventAccessHTTP проверяет штатный роутер, подпись токена и реальные связи PostgreSQL.
func TestEventAccessHTTP(t *testing.T) {
	database := setupTestDB(t)
	t.Cleanup(func() { teardownTestDB(t, database) })
	if err := database.DB.Exec(`
		INSERT INTO users(id,user_id,is_dummy) VALUES(10,100,false),(20,200,false);
		INSERT INTO events(id,name) VALUES(1,'own'),(2,'foreign');
		INSERT INTO user_event(user_id,event_id) VALUES(10,1),(20,2);
		INSERT INTO transactions(id,event_id,name,total_paid) VALUES(1,1,'own',10),(2,2,'foreign',10);
		INSERT INTO tasks(id,event_id,title) VALUES(1,1,'own'),(2,2,'foreign');
		INSERT INTO activities(id,event_id,description) VALUES(1,1,'own'),(2,2,'foreign');
	`).Error; err != nil {
		t.Fatal(err)
	}
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	keys := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, base64.StdEncoding.EncodeToString(pub))
	}))
	t.Cleanup(keys.Close)
	c, err := createTestContainer(t, &config.Config{}, gin.New(), database.DB, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	c.AuthClient = auth.NewClient(keys.URL, time.Hour)
	c.ServerHandler = handler.NewServerHandler(c.EventService, c.UserService, c.TransactionService, c.ActivityService, c.TaskService, c.CategoryService, c.IconService, nil)
	c.RegisterRoutes()
	payload, err := json.Marshal(map[string]interface{}{"user_id": 100, "roles": []string{"user"}, "exp": time.Now().Add(time.Hour).Unix(), "purpose": "access"})
	if err != nil {
		t.Fatal(err)
	}
	token, err := json.Marshal(auth.Token{Payload: payload, Sig: ed25519.Sign(priv, payload)})
	if err != nil {
		t.Fatal(err)
	}
	bearer := "Bearer " + base64.StdEncoding.EncodeToString(token)
	request := func(method, path string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("Authorization", bearer)
		res := httptest.NewRecorder()
		c.Router.ServeHTTP(res, req)
		return res
	}
	for _, resource := range []string{"", "/transaction", "/activity", "/task", "/user", "/debts", "/optimized-debts", "/user/dummies", "/user/dummy"} {
		for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
			path := "/api/v1/event/2" + resource
			// Проверяем лишь зарегистрированные сочетания метода и маршрута.
			found := false
			for _, route := range c.Router.Routes() {
				if route.Method == method && route.Path == "/api/v1/event/:id_event"+resource {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			t.Run("foreign/"+method+resource, func(t *testing.T) {
				res := request(method, path)
				if res.Code != 403 {
					t.Fatalf("status=%d body=%s", res.Code, res.Body)
				}
			})
		}
	}
	for _, resource := range []string{"transaction", "task", "activity"} {
		for _, method := range []string{"GET", "PUT", "DELETE"} {
			t.Run("wrongParent/"+method+resource, func(t *testing.T) {
				res := request(method, "/api/v1/event/1/"+resource+"/2")
				if res.Code != 403 {
					t.Fatalf("status=%d body=%s", res.Code, res.Body)
				}
			})
		}
	}
	for _, path := range []string{"/api/v1/event/1/user/20", "/api/v1/event/1/user/20/optimized-debts"} {
		method := "DELETE"
		if path[len(path)-5:] == "debts" {
			method = "GET"
		}
		res := request(method, path)
		if res.Code != 403 {
			t.Fatalf("%s: status=%d body=%s", path, res.Code, res.Body)
		}
	}
	// Сценарий: участник может читать задачи своего мероприятия и удалять свою задачу.
	if res := request("GET", "/api/v1/event/1/task"); res.Code != 200 {
		t.Fatalf("member read: status=%d body=%s", res.Code, res.Body)
	}
	if res := request("DELETE", "/api/v1/event/1/task/1"); res.Code != 200 {
		t.Fatalf("member delete: status=%d body=%s", res.Code, res.Body)
	}
	for _, table := range []string{"events", "transactions", "activities"} {
		var count int64
		if err := database.DB.Table(table).Where("id = ?", 2).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("foreign %s was changed", table)
		}
	}
	// Сценарии: создатель автоматически включён, даже если members отсутствует или содержит только другого пользователя.
	if err := database.DB.Exec("SELECT setval('events_id_seq', 2)").Error; err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, body string
		members    int64
	}{
		{"noMembers", `{"name":"created"}`, 1},
		{"otherMember", `{"name":"created","members":{"user_ids":[200]}}`, 2},
		{"creatorAlreadyPresent", `{"name":"created","members":{"user_ids":[100]}}`, 1},
	} {
		t.Run("creator/"+tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/v1/event", strings.NewReader(tc.body))
			req.Header.Set("Authorization", bearer)
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()
			c.Router.ServeHTTP(res, req)
			if res.Code != 201 {
				t.Fatalf("create: status=%d body=%s", res.Code, res.Body)
			}
			var created struct {
				ID int64 `json:"id"`
			}
			if err := json.Unmarshal(res.Body.Bytes(), &created); err != nil {
				t.Fatal(err)
			}
			var count int64
			if err := database.DB.Table("user_event").Where("event_id = ? AND user_id = ?", created.ID, 10).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != 1 {
				t.Fatalf("creator membership count=%d want=1", count)
			}
			if err := database.DB.Table("user_event").Where("event_id = ?", created.ID).Count(&count).Error; err != nil {
				t.Fatal(err)
			}
			if count != tc.members {
				t.Fatalf("members=%d want=%d", count, tc.members)
			}
			if access := request("GET", fmt.Sprintf("/api/v1/event/%d/task", created.ID)); access.Code != 200 {
				t.Fatalf("creator cannot access event: status=%d body=%s", access.Code, access.Body)
			}
		})
	}
}

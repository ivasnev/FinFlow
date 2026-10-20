package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
)

// TestDummiesUseInternalIdentity проверяет выборку гостей с NULL-внешним ID и разделение мероприятий.
func TestDummiesUseInternalIdentity(t *testing.T) {
	database := setupTestDB(t)
	t.Cleanup(func() { teardownTestDB(t, database) })
	if err := database.DB.Exec(`INSERT INTO users(id,user_id,is_dummy,name_cashed) VALUES(10,NULL,true,'guest1'),(11,NULL,true,'guest2'),(30,NULL,true,'foreign'),(20,10,false,'real');
		INSERT INTO events(id,name) VALUES(1,'first'),(2,'second'),(3,'empty');
		INSERT INTO user_event(user_id,event_id) VALUES(10,1),(11,1),(20,1),(30,2);`).Error; err != nil {
		t.Fatal(err)
	}
	c, err := createTestContainer(t, &config.Config{}, gin.New(), database.DB, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	h := handler.NewServerHandler(c.EventService, c.UserService, c.TransactionService, c.ActivityService, c.TaskService, c.CategoryService, c.IconService, nil)
	api.RegisterHandlers(c.Router, h)
	for _, tc := range []struct {
		event int
		want  []int64
	}{{1, []int64{10, 11}}, {2, []int64{30}}, {3, []int64{}}} {
		t.Run(fmt.Sprintf("event%d", tc.event), func(t *testing.T) {
			res := httptest.NewRecorder()
			c.Router.ServeHTTP(res, httptest.NewRequest("GET", fmt.Sprintf("/api/v1/event/%d/user/dummies", tc.event), nil))
			if res.Code != 200 {
				t.Fatalf("status=%d body=%s", res.Code, res.Body)
			}
			var body api.UserListResponse
			if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Users == nil {
				t.Fatal("missing users")
			}
			ids := make([]int64, 0, len(*body.Users))
			for _, user := range *body.Users {
				if user.InternalId == nil || user.UserId != nil {
					t.Fatalf("invalid dummy identity: %+v", user)
				}
				ids = append(ids, *user.InternalId)
			}
			sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
			if !reflect.DeepEqual(ids, tc.want) {
				t.Fatalf("ids=%v want=%v", ids, tc.want)
			}
		})
	}
}

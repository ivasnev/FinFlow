package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/handler"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
)

// TestEventUpdateOptionalValues проверяет пропуск, очистку и замену полей через HTTP и настоящий PostgreSQL.
func TestEventUpdateOptionalValues(t *testing.T) {
	database, c := setupOptimizationTest(t)
	h := handler.NewServerHandler(c.EventService, c.UserService, c.TransactionService, c.ActivityService, c.TaskService, c.CategoryService, c.IconService)
	api.RegisterHandlers(c.Router, h)
	if err := database.DB.Exec(`INSERT INTO event_categories(id,name) VALUES(10,'first'),(20,'second');`).Error; err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, body, description string
		category                *int
	}{
		{"omitted", `{"name":"updated"}`, "old description", intPointer(10)},
		{"empty description", `{"name":"updated","description":""}`, "", intPointer(10)},
		{"null description", `{"name":"updated","description":null}`, "", intPointer(10)},
		{"clear category", `{"name":"updated","category_id":null}`, "old description", nil},
		{"clear both", `{"name":"updated","description":"","category_id":null}`, "", nil},
		{"replace both", `{"name":"updated","description":"new description","category_id":20}`, "new description", intPointer(20)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := database.DB.Exec(`UPDATE events SET name='old',description='old description',category_id=10,image_id='photo',status='archive' WHERE id=1`).Error; err != nil {
				t.Fatal(err)
			}
			request := httptest.NewRequest(http.MethodPut, "/api/v1/event/1", strings.NewReader(tc.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			c.Router.ServeHTTP(response, request)
			if response.Code != http.StatusOK {
				t.Fatalf("update: %d %s", response.Code, response.Body)
			}
			var actual api.EventResponse
			if err := json.Unmarshal(response.Body.Bytes(), &actual); err != nil {
				t.Fatal(err)
			}
			stored, err := c.EventRepository.GetByID(context.Background(), 1)
			if err != nil {
				t.Fatal(err)
			}
			if actual.Id == nil || *actual.Id != 1 || actual.Name == nil || *actual.Name != "updated" ||
				actual.Description == nil || *actual.Description != tc.description || !equalOptionalInt(actual.CategoryId, tc.category) ||
				actual.PhotoId == nil || *actual.PhotoId != "photo" {
				t.Fatalf("response differs from requested/preserved values: %+v", actual)
			}
			if stored.Name != "updated" || stored.Description != tc.description || !equalOptionalInt(stored.CategoryID, tc.category) ||
				stored.ImageID != "photo" || stored.Status != "archive" {
				t.Fatalf("stored values differ: %+v", stored)
			}
		})
	}
	// Сценарий: nullable-поля не превращают отсутствующее обязательное имя в запись пустой строки.
	for _, body := range []string{`{"description":""}`, `{"name":null}`, `{"name":""}`} {
		request := httptest.NewRequest(http.MethodPut, "/api/v1/event/1", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		c.Router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("invalid name: %d %s", response.Code, response.Body)
		}
	}
}

// intPointer задаёт присутствующую категорию в таблице сценариев.
func intPointer(value int) *int { return &value }

// equalOptionalInt сравнивает значения категории с учётом отсутствия.
func equalOptionalInt(left, right *int) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

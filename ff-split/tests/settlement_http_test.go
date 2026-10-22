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

	"github.com/ivasnev/FinFlow/ff-auth/pkg/auth"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/handler"
)

// TestSettlementHTTP проверяет реальные подписи, пространства ID и запрет изменения чужой записи.
func TestSettlementHTTP(t *testing.T) {
	_, c := settlementFixture(t)
	settlementExpense(t, c, 1, 2, 100)
	public, private, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	keys := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/auth/introspect" {
			var token auth.Token
			raw, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
			_ = json.Unmarshal(raw, &token)
			var payload auth.TokenPayload
			_ = json.Unmarshal(token.Payload, &payload)
			_ = json.NewEncoder(w).Encode(map[string]any{"active": true, "user_id": payload.UserID})
			return
		}
		fmt.Fprint(w, base64.StdEncoding.EncodeToString(public))
	}))
	t.Cleanup(keys.Close)
	c.AuthClient = auth.NewClient(keys.URL, time.Hour)
	c.ServerHandler = handler.NewServerHandler(c.EventService, c.UserService, c.TransactionService, c.ActivityService, c.TaskService, c.CategoryService, c.IconService, nil)
	c.RegisterRoutes()
	request := func(external int64, method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if external > 0 {
			payload, _ := json.Marshal(map[string]any{"user_id": external, "roles": []string{"user"}, "exp": time.Now().Add(time.Hour).Unix(), "purpose": "access"})
			token, _ := json.Marshal(auth.Token{Payload: payload, Sig: ed25519.Sign(private, payload)})
			req.Header.Set("Authorization", "Bearer "+base64.StdEncoding.EncodeToString(token))
		}
		out := httptest.NewRecorder()
		c.Router.ServeHTTP(out, req)
		return out
	}
	valid := `{"from_user_id":2,"amount_cents":1000,"request_id":"c0c0c0c0-1111-4222-8333-444444444444"}`
	for _, tc := range []struct {
		name               string
		actor              int64
		method, path, body string
		status             int
	}{
		{"unauthenticated", 0, "GET", "/api/v1/event/1/settlements", "", 401},
		{"foreign event", 101, "POST", "/api/v1/event/2/settlements", valid, 403},
		{"forged recipient", 101, "POST", "/api/v1/event/1/settlements", strings.Replace(valid, `"from_user_id":2`, `"from_user_id":2,"to_user_id":3`, 1), 400},
		{"fractional cents", 101, "POST", "/api/v1/event/1/settlements", strings.Replace(valid, "1000", "10.5", 1), 400},
		{"overpayment", 101, "POST", "/api/v1/event/1/settlements", strings.Replace(valid, "1000", "10001", 1), 400},
		{"negative amount", 101, "POST", "/api/v1/event/1/settlements", strings.Replace(valid, "1000", "-1", 1), 400},
		{"not a member", 404, "POST", "/api/v1/event/1/settlements", valid, 403},
		{"nonparticipant sender", 101, "POST", "/api/v1/event/1/settlements", strings.Replace(valid, `"from_user_id":2`, `"from_user_id":999`, 1), 400},
		{"sender cannot confirm receipt", 202, "POST", "/api/v1/event/1/settlements", valid, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := request(tc.actor, tc.method, tc.path, tc.body)
			if strings.Contains(got.Body.String(), "validation error on field") {
				t.Fatal("transport details leaked into user message")
			}
			if got.Code != tc.status {
				t.Fatalf("status=%d want=%d body=%s", got.Code, tc.status, got.Body)
			}
		})
	}
	got := request(101, "POST", "/api/v1/event/1/settlements", valid)
	if got.Code != 200 {
		t.Fatalf("create status=%d body=%s", got.Code, got.Body)
	}
	var row struct {
		ID int64 `json:"id"`
		To int64 `json:"to_user_id"`
	}
	if err := json.Unmarshal(got.Body.Bytes(), &row); err != nil {
		t.Fatal(err)
	}
	if row.To != 1 {
		t.Fatalf("external ID used as internal: %+v", row)
	}
	if got := request(101, "POST", "/api/v1/event/1/settlements", valid); got.Code != 200 {
		t.Fatalf("retry status=%d %s", got.Code, got.Body)
	}
	if got := request(101, "POST", "/api/v1/event/1/settlements", strings.Replace(valid, "1000", "1001", 1)); got.Code != 409 {
		t.Fatalf("conflict status=%d %s", got.Code, got.Body)
	}
	for _, tc := range []struct {
		actor  int64
		event  int
		status int
	}{{202, 1, 403}, {101, 2, 403}, {101, 1, 200}} {
		got := request(tc.actor, "DELETE", fmt.Sprintf("/api/v1/event/%d/settlements/%d", tc.event, row.ID), "")
		if got.Code != tc.status {
			t.Fatalf("cancel status=%d want=%d %s", got.Code, tc.status, got.Body)
		}
	}
	if got := request(101, "GET", "/api/v1/event/1/settlements?before=0", ""); got.Code != 400 {
		t.Fatalf("invalid cursor status=%d", got.Code)
	}
}

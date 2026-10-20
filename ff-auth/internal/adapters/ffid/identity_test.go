package ffid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestRegistrationFailureClassification проверяет ошибки источника после успешной предварительной проверки.
func TestRegistrationFailureClassification(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"conflict", 409, `{}`, ErrConflict},
		{"unavailable", 500, `{}`, ErrUnavailable},
		{"missing body", 201, ``, ErrUnavailable},
		{"wrong identity", 201, `{"id":2}`, ErrUnavailable},
		{"success", 201, `{"id":1,"email":"user@example.com","nickname":"user"}`, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			}))
			t.Cleanup(server.Close)
			client, err := NewAdapter(server.URL, server.Client())
			if err != nil {
				t.Fatal(err)
			}
			_, err = client.RegisterUser(context.Background(), &RegisterUserRequest{UserID: 1, Email: "user@example.com", Nickname: "user"})
			if !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want=%v", err, tc.want)
			}
		})
	}
	// Сценарий: сетевой отказ также сохраняет классификацию недоступности.
	server := httptest.NewServer(http.NotFoundHandler())
	server.Close()
	client, err := NewAdapter(server.URL, http.DefaultClient)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.RegisterUser(context.Background(), &RegisterUserRequest{UserID: 1})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatalf("transport error lost classification: %v", err)
	}
}

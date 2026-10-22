package receiptimages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestExistingBucketPrivacy проверяет отказ для публичного или недоступного для проверки бакета.
func TestExistingBucketPrivacy(t *testing.T) {
	for _, mode := range []string{"private", "public", "unknown"} {
		// Сценарий: проверка отдельного варианта входных данных.
		t.Run(mode, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == "HEAD" {
					w.WriteHeader(200)
					return
				}
				if r.URL.Query().Has("location") {
					w.Header().Set("Content-Type", "application/xml")
					w.Write([]byte(`<LocationConstraint xmlns="http://s3.amazonaws.com/doc/2006-03-01/"></LocationConstraint>`))
					return
				}
				if r.URL.Query().Has("policy") {
					switch mode {
					case "private":
						w.WriteHeader(404)
						w.Write([]byte(`<Error><Code>NoSuchBucketPolicy</Code><Message>no policy</Message></Error>`))
					case "public":
						w.Write([]byte(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":["s3:GetObject"],"Resource":["arn:aws:s3:::private/*"]}]}`))
					case "unknown":
						w.WriteHeader(403)
						w.Write([]byte(`<Error><Code>AccessDenied</Code><Message>denied</Message></Error>`))
					}
					return
				}
				t.Errorf("unexpected S3 request %s", r.URL)
				w.WriteHeader(500)
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			store, err := NewStore(ctx, strings.TrimPrefix(server.URL, "http://"), "synthetic-access", "synthetic-secret", "private", false)
			if mode == "private" {
				if err != nil || store == nil {
					t.Fatalf("private bucket rejected %v", err)
				}
			} else if err == nil || store != nil {
				t.Fatalf("unsafe/unverifiable bucket accepted %s", mode)
			}
		})
	}
}

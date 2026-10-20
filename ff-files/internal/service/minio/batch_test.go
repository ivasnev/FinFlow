package minio

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-files/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
	s3 "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// TestBatchObjectIdentity сверяет возвращаемые ID и ссылки с ключами настоящих S3-запросов SDK.
func TestBatchObjectIdentity(t *testing.T) {
	var mu sync.Mutex
	stored := make(map[string]bool)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			mu.Lock()
			stored[strings.TrimPrefix(r.URL.Path, "/test-bucket/")] = true
			mu.Unlock()
			w.Header().Set("ETag", `"098f6bcd4621d373cade4e832627b4f6"`)
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	client, err := s3.New(strings.TrimPrefix(server.URL, "http://"), &s3.Options{Creds: credentials.NewStaticV4("test", "test", ""), Region: "us-east-1", Secure: false, BucketLookup: s3.BucketLookupPath})
	if err != nil {
		t.Fatal(err)
	}
	s := &minioServiceImpl{mc: client, cfg: &config.MinIO{BucketName: "test-bucket", FileTimeExpiration: 300}}
	got, err := s.CreateMany([]service.FileData{{FileName: "same.txt", Data: []byte("first")}, {FileName: "same.txt", Data: []byte("second")}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("results=%d want=2", len(got))
	}
	mu.Lock()
	defer mu.Unlock()
	if len(stored) != 2 {
		t.Fatalf("stored objects=%d want=2", len(stored))
	}
	for _, file := range got {
		if _, err := uuid.Parse(file.ObjectID); err != nil {
			t.Fatalf("invalid object UUID: %v", err)
		}
		if !stored[file.ObjectID] {
			t.Fatalf("returned unknown object ID: %s", file.ObjectID)
		}
		link, err := url.Parse(file.Link)
		if err != nil {
			t.Fatal(err)
		}
		if link.Query().Get("X-Amz-Expires") != "300" {
			t.Fatalf("expiry must be seconds: %s", link.Query().Get("X-Amz-Expires"))
		}
		if link.Path != "/test-bucket/"+file.ObjectID {
			t.Fatalf("link does not identify object: %s", file.Link)
		}
	}
	if got[0].ObjectID == got[1].ObjectID {
		t.Fatal("duplicate names share an object ID")
	}
}

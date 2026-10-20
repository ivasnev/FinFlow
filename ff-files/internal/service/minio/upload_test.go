package minio

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ivasnev/FinFlow/ff-files/internal/common/config"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
	s3 "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// TestUploadConcurrency проверяет предел параллелизма и отмену до отправки объектов.
func TestUploadConcurrency(t *testing.T) {
	started := make(chan struct{}, service.MaxUploadFiles)
	release := make(chan struct{})
	var active, maximum, calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		n := active.Add(1)
		defer active.Add(-1)
		for old := maximum.Load(); n > old && !maximum.CompareAndSwap(old, n); old = maximum.Load() {
		}
		started <- struct{}{}
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		w.Header().Set("ETag", `"098f6bcd4621d373cade4e832627b4f6"`)
	}))
	t.Cleanup(server.Close)
	t.Cleanup(func() {
		select {
		case <-release:
		default:
			close(release)
		}
	})
	client, err := s3.New(strings.TrimPrefix(server.URL, "http://"), &s3.Options{Creds: credentials.NewStaticV4("test", "test", ""), Region: "us-east-1", BucketLookup: s3.BucketLookupPath})
	if err != nil {
		t.Fatal(err)
	}
	s := &minioServiceImpl{mc: client, cfg: &config.MinIO{BucketName: "test-bucket", FileTimeExpiration: 300}}
	files := make([]service.FileData, service.MaxUploadFiles)
	for i := range files {
		files[i] = service.FileData{FileName: "file", Reader: strings.NewReader("test"), Size: 4}
	}
	done := make(chan error, 1)
	go func() { _, err := s.CreateMany(context.Background(), files); done <- err }()
	for i := 0; i < 4; i++ {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			t.Fatal("uploads did not start")
		}
	}
	if calls.Load() != 4 {
		t.Fatalf("unbounded concurrency: %d", calls.Load())
	}
	close(release)
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("uploads did not finish")
	}
	if maximum.Load() > 4 || calls.Load() != service.MaxUploadFiles {
		t.Fatalf("maximum=%d calls=%d", maximum.Load(), calls.Load())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := s.CreateMany(ctx, files); err == nil {
		t.Fatal("canceled upload accepted")
	}
	if calls.Load() != service.MaxUploadFiles {
		t.Fatal("canceled request reached storage")
	}
}

// TestStorageUploadBounds проверяет сервисные ограничения до вызова SDK и чтения потоков.
func TestStorageUploadBounds(t *testing.T) {
	for _, files := range [][]service.FileData{
		make([]service.FileData, service.MaxUploadFiles+1),
		{{Reader: strings.NewReader(""), Size: service.MaxUploadBytes + 1}},
		{{Reader: strings.NewReader(""), Size: -1}},
		{{Reader: strings.NewReader(""), Size: service.MaxUploadBytes}, {Reader: strings.NewReader(""), Size: 1}},
	} {
		if _, err := (&minioServiceImpl{}).CreateMany(context.Background(), files); err == nil {
			t.Fatal("invalid upload accepted")
		}
	}
}

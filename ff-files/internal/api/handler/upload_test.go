package handler

import (
	"context"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
)

// zeroReader генерирует данные без выделения памяти размером с файл.
type zeroReader struct{}

// Read заполняет только буфер, предоставленный вызывающей стороной.
func (zeroReader) Read(p []byte) (int, error) { clear(p); return len(p), nil }

// countedBody фиксирует объём фактически прочитанного HTTP-тела.
type countedBody struct {
	io.Reader
	read int64
}

// Read учитывает байты, прочитанные multipart-парсером.
func (b *countedBody) Read(p []byte) (int, error) {
	n, err := b.Reader.Read(p)
	b.read += int64(n)
	return n, err
}

// streamingStorage проверяет потоковый контракт во время вызова хранилища.
type streamingStorage struct {
	service.MinIO
	calls int
	bytes int64
	fail  bool
}

// CreateOne читает поток, проверяя отсутствие полной копии файла в Data.
func (s *streamingStorage) CreateOne(ctx context.Context, file service.FileData) (*service.FileUploadResult, error) {
	s.calls++
	if file.Reader == nil || file.Data != nil {
		return nil, fmt.Errorf("upload is not streamed")
	}
	if s.fail {
		return nil, fmt.Errorf("injected storage failure")
	}
	n, err := io.Copy(io.Discard, file.Reader)
	s.bytes += n
	if err != nil || n != file.Size {
		return nil, fmt.Errorf("stream size mismatch: %d/%d: %v", n, file.Size, err)
	}
	return &service.FileUploadResult{ObjectID: uuid.NewString(), Link: "http://files.local/object"}, nil
}

// TestUploadBounds проверяет реальные границы без создания огромных байтовых срезов.
func TestUploadBounds(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		sizes                []int64
		batch, known, fail   bool
		status               int
		wantRead, wantStored int64
	}{
		{"known oversized body", []int64{maxUploadBodyBytes}, false, true, false, 413, 0, 0},
		{"chunked oversized body", []int64{maxUploadBodyBytes}, false, false, false, 413, maxUploadBodyBytes + 1, 0},
		{"oversized file", []int64{service.MaxUploadBytes + 1}, false, false, false, 413, -1, 0},
		{"oversized total", []int64{service.MaxUploadBytes/2 + 1, service.MaxUploadBytes/2 + 1}, true, false, false, 413, -1, 0},
		{"too many files", make([]int64, service.MaxUploadFiles+1), true, false, false, 413, -1, 0},
		{"single extra file", []int64{1, 1}, false, false, false, 413, -1, 0},
		{"spooled success", []int64{2 << 20}, false, false, false, 200, -1, 2 << 20},
		{"spooled storage failure", []int64{2 << 20}, false, false, true, 500, -1, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			temp := t.TempDir()
			t.Setenv("TMPDIR", temp)
			field := "file"
			if tc.batch {
				field = "files"
			}
			var parts []io.Reader
			var length int64
			for i, size := range tc.sizes {
				prefix := fmt.Sprintf("--boundary\r\nContent-Disposition: form-data; name=%q; filename=\"%d.bin\"\r\n\r\n", field, i)
				parts = append(parts, strings.NewReader(prefix), io.LimitReader(zeroReader{}, size), strings.NewReader("\r\n"))
				length += int64(len(prefix)) + size + 2
			}
			parts = append(parts, strings.NewReader("--boundary--\r\n"))
			length += int64(len("--boundary--\r\n"))
			body := &countedBody{Reader: io.MultiReader(parts...)}
			storage := &streamingStorage{fail: tc.fail}
			router := gin.New()
			handler := NewServerHandler(storage)
			if tc.batch {
				router.POST("/upload", handler.UploadFiles)
			} else {
				router.POST("/upload", handler.UploadFile)
			}
			req := httptest.NewRequest("POST", "/upload", io.NopCloser(body))
			req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
			if tc.known {
				req.ContentLength = length
			} else {
				req.ContentLength = -1
			}
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tc.status {
				t.Fatalf("status=%d want=%d: %s", res.Code, tc.status, res.Body)
			}
			if tc.wantRead >= 0 && body.read != tc.wantRead {
				t.Fatalf("read=%d want=%d", body.read, tc.wantRead)
			}
			if body.read > maxUploadBodyBytes+1 {
				t.Fatalf("unbounded read: %d", body.read)
			}
			if storage.bytes != tc.wantStored || (tc.status == 413 && storage.calls != 0) {
				t.Fatalf("storage calls=%d bytes=%d", storage.calls, storage.bytes)
			}
			entries, err := os.ReadDir(temp)
			if err != nil || len(entries) != 0 {
				t.Fatalf("temporary files leaked: %v %v", entries, err)
			}
		})
	}
}

package handler

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
)

// batchStorage фиксирует файлы, переданные обработчиком хранилищу.
type batchStorage struct {
	service.MinIO
	files []service.FileData
}

// CreateMany сохраняет исходные записи для проверки одинаковых имён.
func (s *batchStorage) CreateMany(files []service.FileData) ([]*service.FileUploadResult, error) {
	s.files = files
	results := make([]*service.FileUploadResult, len(files))
	for i := range files {
		results[i] = &service.FileUploadResult{ObjectID: uuid.NewString(), Link: "http://files.local/object"}
	}
	return results, nil
}

// TestBatchPreservesDuplicateNames проверяет два разных файла с одинаковым именем в multipart.
func TestBatchPreservesDuplicateNames(t *testing.T) {
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	for _, content := range []string{"first", "second"} {
		part, err := form.CreateFormFile("files", "same.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	storage := &batchStorage{}
	router := gin.New()
	router.POST("/upload", NewServerHandler(storage).UploadFiles)
	req := httptest.NewRequest("POST", "/upload", &body)
	req.Header.Set("Content-Type", form.FormDataContentType())
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	if res.Code != 200 {
		t.Fatalf("status=%d body=%s", res.Code, res.Body)
	}
	if len(storage.files) != 2 || string(storage.files[0].Data) != "first" || string(storage.files[1].Data) != "second" {
		t.Fatalf("files were collapsed or reordered: %+v", storage.files)
	}
}

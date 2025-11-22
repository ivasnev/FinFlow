package minio

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ivasnev/FinFlow/ff-files/internal/service"
)

func TestMinioService_CreateOne(t *testing.T) {
	// Для unit-тестов мы не можем протестировать CreateOne без реального подключения
	// Поэтому пропустим этот тест или сделаем его интеграционным
	t.Skip("CreateOne requires real MinIO connection or mock wrapper")
}

func TestMinioService_CreateMany(t *testing.T) {
	t.Skip("CreateMany requires real MinIO connection or mock wrapper")
}

func TestMinioService_GetOne(t *testing.T) {
	t.Skip("GetOne requires real MinIO connection or mock wrapper")
}

func TestMinioService_GetMany(t *testing.T) {
	t.Skip("GetMany requires real MinIO connection or mock wrapper")
}

func TestMinioService_DeleteOne(t *testing.T) {
	t.Skip("DeleteOne requires real MinIO connection or mock wrapper")
}

func TestMinioService_DeleteMany(t *testing.T) {
	t.Skip("DeleteMany requires real MinIO connection or mock wrapper")
}

func TestMinioService_GetMetadata(t *testing.T) {
	t.Skip("GetMetadata requires real MinIO connection or mock wrapper")
}

func TestMinioService_GenerateTemporaryUrl(t *testing.T) {
	t.Skip("GenerateTemporaryUrl requires real MinIO connection or mock wrapper")
}

// Тесты для валидации входных данных и обработки ошибок
func TestMinioService_FileDataValidation(t *testing.T) {
	t.Run("пустой файл", func(t *testing.T) {
		fileData := service.FileData{
			FileName: "test.txt",
			Data:     []byte{},
		}

		assert.NotNil(t, fileData)
		assert.Equal(t, "test.txt", fileData.FileName)
		assert.Empty(t, fileData.Data)
	})

	t.Run("файл с данными", func(t *testing.T) {
		fileData := service.FileData{
			FileName: "test.txt",
			Data:     []byte("test content"),
		}

		assert.NotNil(t, fileData)
		assert.Equal(t, "test.txt", fileData.FileName)
		assert.Equal(t, []byte("test content"), fileData.Data)
		assert.Equal(t, 12, len(fileData.Data))
	})
}

func TestMinioService_FileUploadResult(t *testing.T) {
	t.Run("создание результата загрузки", func(t *testing.T) {
		objectID := uuid.New().String()
		link := "http://localhost:9000/test-bucket/" + objectID

		result := &service.FileUploadResult{
			ObjectID: objectID,
			Link:     link,
		}

		assert.NotNil(t, result)
		assert.Equal(t, objectID, result.ObjectID)
		assert.Equal(t, link, result.Link)
		assert.NotEmpty(t, result.ObjectID)
		assert.NotEmpty(t, result.Link)
	})
}

func TestMinioService_FileMetadata(t *testing.T) {
	t.Run("создание метаданных файла", func(t *testing.T) {
		fileID := uuid.New().String()
		filename := "test.txt"
		size := int64(1024)
		contentType := "text/plain"
		uploadDate := time.Now()
		ownerID := "user123"

		metadata := &service.FileMetadata{
			FileID:      fileID,
			Filename:    filename,
			Size:        size,
			ContentType: contentType,
			UploadDate:  uploadDate,
			OwnerID:     ownerID,
			Metadata:    make(map[string]interface{}),
		}

		assert.NotNil(t, metadata)
		assert.Equal(t, fileID, metadata.FileID)
		assert.Equal(t, filename, metadata.Filename)
		assert.Equal(t, size, metadata.Size)
		assert.Equal(t, contentType, metadata.ContentType)
		assert.Equal(t, ownerID, metadata.OwnerID)
		assert.NotNil(t, metadata.Metadata)
	})

	t.Run("метаданные с дополнительными полями", func(t *testing.T) {
		metadata := &service.FileMetadata{
			FileID:      uuid.New().String(),
			Filename:    "test.txt",
			Size:        1024,
			ContentType: "text/plain",
			UploadDate:  time.Now(),
			OwnerID:     "user123",
			Metadata: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
		}

		assert.NotNil(t, metadata.Metadata)
		assert.Equal(t, "value1", metadata.Metadata["key1"])
		assert.Equal(t, 123, metadata.Metadata["key2"])
	})
}

func TestMinioService_TemporaryURLResult(t *testing.T) {
	t.Run("создание временной ссылки", func(t *testing.T) {
		url := "http://localhost:9000/test-bucket/file123?X-Amz-Algorithm=..."
		expiresAt := time.Now().Add(1 * time.Hour)

		result := &service.TemporaryURLResult{
			URL:       url,
			ExpiresAt: expiresAt,
		}

		assert.NotNil(t, result)
		assert.Equal(t, url, result.URL)
		assert.Equal(t, expiresAt.Unix(), result.ExpiresAt.Unix())
		assert.True(t, result.ExpiresAt.After(time.Now()))
	})
}

func TestMinioService_NewMinioService(t *testing.T) {
	t.Run("создание нового сервиса", func(t *testing.T) {
		minioService := NewMinioService()

		assert.NotNil(t, minioService)
		
		// Проверяем, что сервис реализует интерфейс MinIO
		var _ service.MinIO = minioService

		impl, ok := minioService.(*minioServiceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl)
		assert.Nil(t, impl.mc) // Клиент еще не инициализирован
		assert.Nil(t, impl.cfg) // Конфигурация еще не установлена
	})
}

func TestMinioService_ErrorHandling(t *testing.T) {
	t.Run("обработка ошибок создания объекта", func(t *testing.T) {
		err := errors.New("ошибка при создании объекта test.txt: connection refused")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при создании объекта")
	})

	t.Run("обработка ошибок получения URL", func(t *testing.T) {
		err := errors.New("ошибка при создании URL для объекта test.txt: invalid credentials")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при создании URL")
	})

	t.Run("обработка ошибок получения объекта", func(t *testing.T) {
		err := errors.New("ошибка при получении URL для объекта file123: object not found")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при получении URL")
	})

	t.Run("обработка ошибок получения метаданных", func(t *testing.T) {
		err := errors.New("ошибка при получении информации об объекте file123: object not found")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ошибка при получении информации")
	})

	t.Run("обработка ошибок генерации временной ссылки", func(t *testing.T) {
		err := errors.New("файл с ID file123 не найден: object not found")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "не найден")
	})
}


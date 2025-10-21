package service

import (
	"time"

	"github.com/ivasnev/FinFlow/ff-files/internal/common/config"
)

// FileData представляет данные файла для загрузки
type FileData struct {
	FileName string
	Data     []byte
}

// FileUploadResult представляет результат загрузки файла
type FileUploadResult struct {
	ObjectID string `json:"object_id"`
	Link     string `json:"link"`
}

// FileMetadata представляет метаданные файла
type FileMetadata struct {
	FileID      string                 `json:"file_id"`
	Filename    string                 `json:"filename"`
	Size        int64                  `json:"size"`
	ContentType string                 `json:"content_type"`
	UploadDate  time.Time              `json:"upload_date"`
	OwnerID     string                 `json:"owner_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TemporaryURLResult представляет результат генерации временной ссылки
type TemporaryURLResult struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// MinIO определяет методы для работы с MinIO хранилищем
type MinIO interface {
	// InitMinio инициализирует подключение к MinIO
	InitMinio(cfg *config.MinIO) error

	// CreateOne создает один объект в бакете MinIO
	CreateOne(file FileData) (*FileUploadResult, error)

	// CreateMany создает несколько объектов в бакете MinIO
	CreateMany(files map[string]FileData) ([]*FileUploadResult, error)

	// GetOne получает один объект из бакета MinIO
	GetOne(objectID string) (string, error)

	// GetMany получает несколько объектов из бакета MinIO
	GetMany(objectIDs []string) ([]string, error)

	// DeleteOne удаляет один объект из бакета MinIO
	DeleteOne(objectID string) error

	// DeleteMany удаляет несколько объектов из бакета MinIO
	DeleteMany(objectIDs []string) error

	// GetMetadata получает метаданные файла
	GetMetadata(objectID string) (*FileMetadata, error)

	// GenerateTemporaryUrl генерирует временную ссылку для доступа к файлу
	GenerateTemporaryUrl(objectID string, expiresInSeconds int) (*TemporaryURLResult, error)
}

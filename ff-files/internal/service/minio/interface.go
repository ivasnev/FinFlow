package minio

import (
	"github.com/ivasnev/FinFlow/ff-files/internal/common/config"
	helpers2 "github.com/ivasnev/FinFlow/ff-files/internal/service/minio/helpers"
)

// MinioServiceInterface интерфейс для взаимодействия с Minio
type MinioServiceInterface interface {
	InitMinio(cfg *config.MinIO) error                                              // Метод для инициализации подключения к Minio
	CreateOne(file helpers2.FileDataType) (*helpers2.CreatedObject, error)          // Метод для создания одного объекта в бакете Minio
	CreateMany(map[string]helpers2.FileDataType) ([]*helpers2.CreatedObject, error) // Метод для создания нескольких объектов в бакете Minio
	GetOne(objectID string) (string, error)                                         // Метод для получения одного объекта из бакета Minio
	GetMany(objectIDs []string) ([]string, error)                                   // Метод для получения нескольких объектов из бакета Minio
	DeleteOne(objectID string) error                                                // Метод для удаления одного объекта из бакета Minio
	DeleteMany(objectIDs []string) error                                            // Метод для удаления нескольких объектов из бакета Minio
}

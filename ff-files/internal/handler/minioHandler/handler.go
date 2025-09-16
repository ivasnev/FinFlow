package minioHandler

import (
	"FinFlow/ff-files/internal/service/minio"
)

type Handler struct {
	minioService minio.MinioServiceInterface
}

func NewMinioHandler(
	minioService minio.MinioServiceInterface,
) *Handler {
	return &Handler{
		minioService: minioService,
	}
}

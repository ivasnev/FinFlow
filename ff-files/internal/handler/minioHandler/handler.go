package minioHandler

import (
	"github.com/ivasnev/FinFlow/ff-files/internal/service/minio"
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

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-files/internal/config"
	"github.com/ivasnev/FinFlow/ff-files/internal/model"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gorm.io/gorm"
)

type Service struct {
	db     *gorm.DB
	minio  *minio.Client
	config *config.Config
}

func NewService(db *gorm.DB, cfg *config.Config) (*Service, error) {
	minioClient, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKeyID, cfg.MinIO.SecretAccessKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &Service{
		db:     db,
		minio:  minioClient,
		config: cfg,
	}, nil
}

func (s *Service) UploadFile(ctx context.Context, reader io.Reader, size int64, mimeType string, ownerID string, metadata map[string]interface{}) (*model.File, error) {
	fileID := uuid.New()
	storagePath := filepath.Join(ownerID, fileID.String())

	// Загружаем файл в MinIO
	_, err := s.minio.PutObject(ctx, s.config.MinIO.BucketName, storagePath, reader, size, minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to storage: %w", err)
	}

	// Преобразуем метаданные в JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Создаем запись в БД
	file := &model.File{
		ID:          fileID,
		Size:        size,
		MimeType:    mimeType,
		OwnerID:     ownerID,
		UploadedAt:  time.Now(),
		Metadata:    string(metadataJSON),
		StoragePath: storagePath,
	}

	if err := s.db.Create(file).Error; err != nil {
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	return file, nil
}

func (s *Service) GetFile(ctx context.Context, fileID uuid.UUID) (*model.File, error) {
	var file model.File
	if err := s.db.First(&file, "id = ?", fileID).Error; err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	return &file, nil
}

func (s *Service) DeleteFile(ctx context.Context, fileID uuid.UUID) error {
	file, err := s.GetFile(ctx, fileID)
	if err != nil {
		return err
	}

	// Удаляем файл из хранилища
	if err := s.minio.RemoveObject(ctx, s.config.MinIO.BucketName, file.StoragePath, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to delete file from storage: %w", err)
	}

	// Удаляем запись из БД
	if err := s.db.Delete(file).Error; err != nil {
		return fmt.Errorf("failed to delete file metadata: %w", err)
	}

	return nil
}

func (s *Service) GetFileMetadata(ctx context.Context, fileID uuid.UUID) (*model.File, error) {
	return s.GetFile(ctx, fileID)
}

func (s *Service) GenerateTemporaryURL(ctx context.Context, fileID uuid.UUID, expiresIn time.Duration) (string, error) {
	file, err := s.GetFile(ctx, fileID)
	if err != nil {
		return "", err
	}

	url, err := s.minio.PresignedGetObject(ctx, s.config.MinIO.BucketName, file.StoragePath, expiresIn, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate temporary URL: %w", err)
	}

	return url.String(), nil
}

func (s *Service) GetObject(ctx context.Context, storagePath string) (*minio.Object, error) {
	return s.minio.GetObject(ctx, s.config.MinIO.BucketName, storagePath, minio.GetObjectOptions{})
}

package service

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/ivasnev/FinFlow/ff-files/internal/models"
	"github.com/ivasnev/FinFlow/ff-files/internal/repository"
	"github.com/ivasnev/FinFlow/ff-files/internal/storage"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"
)

type FileService interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, uploader string) (*models.File, error)
	UploadFileFromURL(ctx context.Context, url string, uploader string) (*models.File, error)
	GetFile(ctx context.Context, id string) (*models.File, io.ReadCloser, error)
	GetFileMetadata(ctx context.Context, id string) (*models.File, []models.FileMetadata, error)
	DeleteFile(ctx context.Context, id string) error
	GenerateTemporaryURL(ctx context.Context, fileID string, duration time.Duration) (*models.TemporaryURL, error)
	GetFileByTemporaryURL(ctx context.Context, urlID string) (*models.File, io.ReadCloser, error)
	CleanupExpiredFiles(ctx context.Context) error
	CleanupExpiredURLs(ctx context.Context) error
}

type fileService struct {
	repo    repository.FileRepository
	storage storage.Storage
	config  *Config
}

type Config struct {
	MaxFileSize      int64
	AllowedMimeTypes []string
	SoftDeleteTimeout time.Duration
}

func NewFileService(repo repository.FileRepository, storage storage.Storage, config *Config) FileService {
	return &fileService{
		repo:    repo,
		storage: storage,
		config:  config,
	}
}

func (s *fileService) validateFile(size int64, mimeType string) error {
	if size > s.config.MaxFileSize {
		return fmt.Errorf("file size exceeds maximum allowed size of %d bytes", s.config.MaxFileSize)
	}

	allowed := false
	for _, mt := range s.config.AllowedMimeTypes {
		if mt == mimeType {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("mime type %s is not allowed", mimeType)
	}
	return nil
}

func (s *fileService) UploadFile(ctx context.Context, fileHeader *multipart.FileHeader, uploader string) (*models.File, error) {
	if err := s.validateFile(fileHeader.Size, fileHeader.Header.Get("Content-Type")); err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileID := uuid.New().String()
	ext := filepath.Ext(fileHeader.Filename)
	path, err := s.storage.Save(fileID+ext, file)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	fileModel := &models.File{
		ID:         fileID,
		Name:       fileHeader.Filename,
		Path:       path,
		Size:       fileHeader.Size,
		MimeType:   fileHeader.Header.Get("Content-Type"),
		UploadedBy: uploader,
		IsDeleted:  false,
	}

	if err := s.repo.CreateFile(ctx, fileModel); err != nil {
		_ = s.storage.Delete(path)
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	return fileModel, nil
}

func (s *fileService) UploadFileFromURL(ctx context.Context, url string, uploader string) (*models.File, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from URL: %w", err)
	}
	defer resp.Body.Close()

	if err := s.validateFile(resp.ContentLength, resp.Header.Get("Content-Type")); err != nil {
		return nil, err
	}

	fileID := uuid.New().String()
	ext := filepath.Ext(url)
	path, err := s.storage.Save(fileID+ext, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	fileModel := &models.File{
		ID:         fileID,
		Name:       filepath.Base(url),
		Path:       path,
		Size:       resp.ContentLength,
		MimeType:   resp.Header.Get("Content-Type"),
		UploadedBy: uploader,
		IsDeleted:  false,
	}

	if err := s.repo.CreateFile(ctx, fileModel); err != nil {
		_ = s.storage.Delete(path)
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	return fileModel, nil
}

func (s *fileService) GetFile(ctx context.Context, id string) (*models.File, io.ReadCloser, error) {
	file, err := s.repo.GetFileByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file record: %w", err)
	}

	if file.IsDeleted {
		return nil, nil, fmt.Errorf("file is deleted")
	}

	reader, err := s.storage.Get(file.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	return file, reader, nil
}

func (s *fileService) GetFileMetadata(ctx context.Context, id string) (*models.File, []models.FileMetadata, error) {
	file, err := s.repo.GetFileByID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file record: %w", err)
	}

	metadata, err := s.repo.GetMetadataByFileID(ctx, id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return file, metadata, nil
}

func (s *fileService) DeleteFile(ctx context.Context, id string) error {
	if err := s.repo.SoftDeleteFile(ctx, id); err != nil {
		return fmt.Errorf("failed to soft delete file: %w", err)
	}
	return nil
}

func (s *fileService) GenerateTemporaryURL(ctx context.Context, fileID string, duration time.Duration) (*models.TemporaryURL, error) {
	file, err := s.repo.GetFileByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file record: %w", err)
	}

	if file.IsDeleted {
		return nil, fmt.Errorf("file is deleted")
	}

	tempURL := &models.TemporaryURL{
		ID:        uuid.New().String(),
		FileID:    fileID,
		ExpiresAt: time.Now().Add(duration),
	}

	if err := s.repo.CreateTemporaryURL(ctx, tempURL); err != nil {
		return nil, fmt.Errorf("failed to create temporary URL: %w", err)
	}

	return tempURL, nil
}

func (s *fileService) GetFileByTemporaryURL(ctx context.Context, urlID string) (*models.File, io.ReadCloser, error) {
	tempURL, err := s.repo.GetTemporaryURL(ctx, urlID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get temporary URL: %w", err)
	}

	return s.GetFile(ctx, tempURL.FileID)
}

func (s *fileService) CleanupExpiredFiles(ctx context.Context) error {
	files, err := s.repo.GetFilesToDelete(ctx, time.Now().Add(-s.config.SoftDeleteTimeout))
	if err != nil {
		return fmt.Errorf("failed to get files to delete: %w", err)
	}

	for _, file := range files {
		if err := s.storage.Delete(file.Path); err != nil {
			return fmt.Errorf("failed to delete file from storage: %w", err)
		}

		if err := s.repo.HardDeleteFile(ctx, file.ID); err != nil {
			return fmt.Errorf("failed to hard delete file record: %w", err)
		}
	}

	return nil
}

func (s *fileService) CleanupExpiredURLs(ctx context.Context) error {
	return s.repo.DeleteExpiredTemporaryURLs(ctx)
} 
package repository

import (
	"context"
	"github.com/ivasnev/FinFlow/ff-files/internal/models"
	"gorm.io/gorm"
	"time"
)

type FileRepository interface {
	CreateFile(ctx context.Context, file *models.File) error
	GetFileByID(ctx context.Context, id string) (*models.File, error)
	UpdateFile(ctx context.Context, file *models.File) error
	SoftDeleteFile(ctx context.Context, id string) error
	HardDeleteFile(ctx context.Context, id string) error
	GetFilesToDelete(ctx context.Context, olderThan time.Time) ([]models.File, error)

	CreateMetadata(ctx context.Context, metadata *models.FileMetadata) error
	GetMetadataByFileID(ctx context.Context, fileID string) ([]models.FileMetadata, error)
	
	CreateTemporaryURL(ctx context.Context, tempURL *models.TemporaryURL) error
	GetTemporaryURL(ctx context.Context, id string) (*models.TemporaryURL, error)
	DeleteExpiredTemporaryURLs(ctx context.Context) error
}

type fileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) CreateFile(ctx context.Context, file *models.File) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *fileRepository) GetFileByID(ctx context.Context, id string) (*models.File, error) {
	var file models.File
	err := r.db.WithContext(ctx).First(&file, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *fileRepository) UpdateFile(ctx context.Context, file *models.File) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *fileRepository) SoftDeleteFile(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&models.File{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_deleted":  true,
			"deleted_at": &now,
		}).Error
}

func (r *fileRepository) HardDeleteFile(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.File{}, "id = ?", id).Error
}

func (r *fileRepository) GetFilesToDelete(ctx context.Context, olderThan time.Time) ([]models.File, error) {
	var files []models.File
	err := r.db.WithContext(ctx).
		Where("is_deleted = ? AND deleted_at < ?", true, olderThan).
		Find(&files).Error
	return files, err
}

func (r *fileRepository) CreateMetadata(ctx context.Context, metadata *models.FileMetadata) error {
	return r.db.WithContext(ctx).Create(metadata).Error
}

func (r *fileRepository) GetMetadataByFileID(ctx context.Context, fileID string) ([]models.FileMetadata, error) {
	var metadata []models.FileMetadata
	err := r.db.WithContext(ctx).
		Where("file_id = ?", fileID).
		Find(&metadata).Error
	return metadata, err
}

func (r *fileRepository) CreateTemporaryURL(ctx context.Context, tempURL *models.TemporaryURL) error {
	return r.db.WithContext(ctx).Create(tempURL).Error
}

func (r *fileRepository) GetTemporaryURL(ctx context.Context, id string) (*models.TemporaryURL, error) {
	var tempURL models.TemporaryURL
	err := r.db.WithContext(ctx).
		Where("id = ? AND expires_at > ?", id, time.Now()).
		First(&tempURL).Error
	if err != nil {
		return nil, err
	}
	return &tempURL, nil
}

func (r *fileRepository) DeleteExpiredTemporaryURLs(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.TemporaryURL{}).Error
} 
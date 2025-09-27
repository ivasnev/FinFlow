package postgres

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// CategoryRepository представляет собой репозиторий для работы с категориями в PostgreSQL
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository создает новый экземпляр CategoryRepository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

// GetCategories возвращает все категории
func (r *CategoryRepository) GetCategories(ctx context.Context) ([]models.Category, error) {
	var categories []models.Category
	if err := r.db.WithContext(ctx).Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// GetCategoryByID возвращает категорию по ID
func (r *CategoryRepository) GetCategoryByID(ctx context.Context, id int) (models.Category, error) {
	var category models.Category
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		return models.Category{}, err
	}
	return category, nil
}

// CreateCategory создает новую категорию
func (r *CategoryRepository) CreateCategory(ctx context.Context, category models.Category) (models.Category, error) {
	if err := r.db.WithContext(ctx).Create(&category).Error; err != nil {
		return models.Category{}, err
	}
	return category, nil
}

// UpdateCategory обновляет существующую категорию
func (r *CategoryRepository) UpdateCategory(ctx context.Context, category models.Category) error {
	return r.db.WithContext(ctx).Save(&category).Error
}

// DeleteCategory удаляет категорию
func (r *CategoryRepository) DeleteCategory(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Delete(&models.Category{}, id).Error
}

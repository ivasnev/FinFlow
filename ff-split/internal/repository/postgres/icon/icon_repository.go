package icon

import (
	"errors"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// IconRepository интерфейс репозитория иконок
type IconRepository struct {
	db *gorm.DB
}

// NewIconRepository создает новый репозиторий иконок
func NewIconRepository(db *gorm.DB) *IconRepository {
	return &IconRepository{db: db}
}

// GetIcons возвращает список всех иконок
func (r *IconRepository) GetIcons() ([]models.Icon, error) {
	var icons []models.Icon
	if err := r.db.Find(&icons).Error; err != nil {
		return nil, err
	}
	return icons, nil
}

// GetIconByID возвращает иконку по идентификатору
func (r *IconRepository) GetIconByID(id uint) (*models.Icon, error) {
	var icon models.Icon
	if err := r.db.First(&icon, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("иконка не найдена")
		}
		return nil, err
	}
	return &icon, nil
}

// CreateIcon создает новую иконку
func (r *IconRepository) CreateIcon(icon *models.Icon) error {
	return r.db.Create(icon).Error
}

// UpdateIcon обновляет существующую иконку
func (r *IconRepository) UpdateIcon(icon *models.Icon) error {
	result := r.db.Save(icon)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("иконка не найдена")
	}
	return nil
}

// DeleteIcon удаляет иконку по идентификатору
func (r *IconRepository) DeleteIcon(id uint) error {
	result := r.db.Delete(&models.Icon{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("иконка не найдена")
	}
	return nil
}


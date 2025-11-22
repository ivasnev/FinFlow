package icon

import (
	"errors"
	"strconv"

	customErrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
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
	var dbIcons []Icon
	if err := r.db.Find(&dbIcons).Error; err != nil {
		return nil, err
	}
	return extractSlice(dbIcons), nil
}

// GetIconByID возвращает иконку по идентификатору
func (r *IconRepository) GetIconByID(id uint) (*models.Icon, error) {
	var dbIcon Icon
	if err := r.db.First(&dbIcon, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customErrors.NewEntityNotFoundError(strconv.Itoa(int(id)), "icon")
		}
		return nil, err
	}
	return extract(&dbIcon), nil
}

// CreateIcon создает новую иконку
func (r *IconRepository) CreateIcon(icon *models.Icon) error {
	dbIcon := load(icon)
	if err := r.db.Create(dbIcon).Error; err != nil {
		return err
	}
	icon.ID = dbIcon.ID
	return nil
}

// UpdateIcon обновляет существующую иконку
func (r *IconRepository) UpdateIcon(icon *models.Icon) error {
	dbIcon := load(icon)
	result := r.db.Save(dbIcon)
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
	result := r.db.Delete(&Icon{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("иконка не найдена")
	}
	return nil
}

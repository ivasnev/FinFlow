package icon

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// extract преобразует модель иконки БД в бизнес-модель
func extract(dbIcon *Icon) *models.Icon {
	if dbIcon == nil {
		return nil
	}

	return &models.Icon{
		ID:       dbIcon.ID,
		Name:     dbIcon.Name,
		FileUUID: dbIcon.FileUUID,
	}
}

// extractSlice преобразует слайс моделей иконок БД в бизнес-модели
func extractSlice(dbIcons []Icon) []models.Icon {
	icons := make([]models.Icon, len(dbIcons))
	for i, dbIcon := range dbIcons {
		if extracted := extract(&dbIcon); extracted != nil {
			icons[i] = *extracted
		}
	}
	return icons
}

// load преобразует бизнес-модель иконки в модель БД
func load(icon *models.Icon) *Icon {
	if icon == nil {
		return nil
	}

	return &Icon{
		ID:       icon.ID,
		Name:     icon.Name,
		FileUUID: icon.FileUUID,
	}
}

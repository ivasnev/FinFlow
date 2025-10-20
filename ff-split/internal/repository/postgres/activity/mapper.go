package activity

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// extract преобразует модель активности БД в бизнес-модель
func extract(dbActivity *Activity) *models.Activity {
	if dbActivity == nil {
		return nil
	}

	return &models.Activity{
		ID:          dbActivity.ID,
		EventID:     dbActivity.EventID,
		UserID:      dbActivity.UserID,
		Description: dbActivity.Description,
		IconID:      dbActivity.IconID,
		CreatedAt:   dbActivity.CreatedAt,
	}
}

// extractSlice преобразует слайс моделей активностей БД в бизнес-модели
func extractSlice(dbActivities []Activity) []models.Activity {
	activities := make([]models.Activity, len(dbActivities))
	for i, dbActivity := range dbActivities {
		if extracted := extract(&dbActivity); extracted != nil {
			activities[i] = *extracted
		}
	}
	return activities
}

// load преобразует бизнес-модель активности в модель БД
func load(activity *models.Activity) *Activity {
	if activity == nil {
		return nil
	}

	return &Activity{
		ID:          activity.ID,
		EventID:     activity.EventID,
		UserID:      activity.UserID,
		Description: activity.Description,
		IconID:      activity.IconID,
		CreatedAt:   activity.CreatedAt,
	}
}

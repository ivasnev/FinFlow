package event

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// extract преобразует модель мероприятия БД в бизнес-модель
func extract(dbEvent *Event) *models.Event {
	if dbEvent == nil {
		return nil
	}

	return &models.Event{
		ID:          dbEvent.ID,
		Name:        dbEvent.Name,
		Description: dbEvent.Description,
		CategoryID:  dbEvent.CategoryID,
		ImageID:     dbEvent.ImageID,
		Status:      dbEvent.Status,
	}
}

// extractSlice преобразует слайс моделей мероприятий БД в бизнес-модели
func extractSlice(dbEvents []Event) []models.Event {
	events := make([]models.Event, len(dbEvents))
	for i, dbEvent := range dbEvents {
		if extracted := extract(&dbEvent); extracted != nil {
			events[i] = *extracted
		}
	}
	return events
}

// load преобразует бизнес-модель мероприятия в модель БД
func load(event *models.Event) *Event {
	if event == nil {
		return nil
	}

	return &Event{
		ID:          event.ID,
		Name:        event.Name,
		Description: event.Description,
		CategoryID:  event.CategoryID,
		ImageID:     event.ImageID,
		Status:      event.Status,
	}
}

// extractEventCategory преобразует модель категории мероприятия БД в бизнес-модель
func extractEventCategory(dbCategory *EventCategory) *models.EventCategory {
	if dbCategory == nil {
		return nil
	}

	return &models.EventCategory{
		ID:     dbCategory.ID,
		Name:   dbCategory.Name,
		IconID: dbCategory.IconID,
	}
}

// loadEventCategory преобразует бизнес-модель категории мероприятия в модель БД
func loadEventCategory(category *models.EventCategory) *EventCategory {
	if category == nil {
		return nil
	}

	return &EventCategory{
		ID:     category.ID,
		Name:   category.Name,
		IconID: category.IconID,
	}
}

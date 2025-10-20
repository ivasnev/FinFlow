package category

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

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

// extractEventCategorySlice преобразует слайс моделей категорий мероприятий БД в бизнес-модели
func extractEventCategorySlice(dbCategories []EventCategory) []models.EventCategory {
	categories := make([]models.EventCategory, len(dbCategories))
	for i, dbCategory := range dbCategories {
		if extracted := extractEventCategory(&dbCategory); extracted != nil {
			categories[i] = *extracted
		}
	}
	return categories
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

// extractTransactionCategory преобразует модель категории транзакции БД в бизнес-модель
func extractTransactionCategory(dbCategory *TransactionCategory) *models.TransactionCategory {
	if dbCategory == nil {
		return nil
	}

	return &models.TransactionCategory{
		ID:     dbCategory.ID,
		Name:   dbCategory.Name,
		IconID: dbCategory.IconID,
	}
}

// extractTransactionCategorySlice преобразует слайс моделей категорий транзакций БД в бизнес-модели
func extractTransactionCategorySlice(dbCategories []TransactionCategory) []models.TransactionCategory {
	categories := make([]models.TransactionCategory, len(dbCategories))
	for i, dbCategory := range dbCategories {
		if extracted := extractTransactionCategory(&dbCategory); extracted != nil {
			categories[i] = *extracted
		}
	}
	return categories
}

// loadTransactionCategory преобразует бизнес-модель категории транзакции в модель БД
func loadTransactionCategory(category *models.TransactionCategory) *TransactionCategory {
	if category == nil {
		return nil
	}

	return &TransactionCategory{
		ID:     category.ID,
		Name:   category.Name,
		IconID: category.IconID,
	}
}

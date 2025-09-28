package postgres

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// CategoryRepository реализует интерфейс repository.CategoryRepository
type CategoryRepository struct {
	db *gorm.DB
	// Стратегии для разных типов категорий
	strategies map[string]CategoryStrategy
}

// CategoryStrategy интерфейс стратегии для работы с конкретным типом категорий
type CategoryStrategy interface {
	GetModel() interface{}
	GetTableName() string
	ConvertToEventCategory(model interface{}) (*models.EventCategory, error)
	ConvertFromEventCategory(category *models.EventCategory) (interface{}, error)
}

// NewCategoryRepository создает новый экземпляр CategoryRepository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	repo := &CategoryRepository{
		db:         db,
		strategies: make(map[string]CategoryStrategy),
	}

	// Регистрация стратегий для разных типов категорий
	repo.strategies["event"] = &EventCategoryStrategy{}
	repo.strategies["transaction"] = &TransactionCategoryStrategy{}

	return repo
}

// GetAll возвращает все категории указанного типа
func (r *CategoryRepository) GetAll(ctx context.Context, categoryType string) ([]models.EventCategory, error) {
	strategy, ok := r.strategies[categoryType]
	if !ok {
		return nil, fmt.Errorf("неизвестный тип категории: %s", categoryType)
	}

	// Создаем слайс правильного типа через reflect
	modelType := reflect.TypeOf(strategy.GetModel()).Elem()
	sliceType := reflect.SliceOf(modelType)
	modelsSlice := reflect.New(sliceType).Interface()

	// Получаем данные из базы
	if err := r.db.WithContext(ctx).Table(strategy.GetTableName()).Find(modelsSlice).Error; err != nil {
		return nil, err
	}

	// Получаем слайс из reflect.Value
	slice := reflect.ValueOf(modelsSlice).Elem()
	length := slice.Len()

	// Конвертируем в общий формат EventCategory
	result := make([]models.EventCategory, 0, length)
	for i := 0; i < length; i++ {
		item := slice.Index(i).Interface()
		category, err := strategy.ConvertToEventCategory(item)
		if err != nil {
			return nil, err
		}
		result = append(result, *category)
	}

	return result, nil
}

// GetByID возвращает категорию по ID
func (r *CategoryRepository) GetByID(ctx context.Context, id int, categoryType string) (*models.EventCategory, error) {
	strategy, ok := r.strategies[categoryType]
	if !ok {
		return nil, fmt.Errorf("неизвестный тип категории: %s", categoryType)
	}

	model := strategy.GetModel()
	if err := r.db.WithContext(ctx).Table(strategy.GetTableName()).First(model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // возвращаем nil, nil если категория не найдена
		}
		return nil, err
	}

	return strategy.ConvertToEventCategory(model)
}

// Create создает новую категорию
func (r *CategoryRepository) Create(ctx context.Context, category *models.EventCategory, categoryType string) (*models.EventCategory, error) {
	strategy, ok := r.strategies[categoryType]
	if !ok {
		return nil, fmt.Errorf("неизвестный тип категории: %s", categoryType)
	}

	model, err := strategy.ConvertFromEventCategory(category)
	if err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Table(strategy.GetTableName()).Create(model).Error; err != nil {
		return nil, err
	}

	return strategy.ConvertToEventCategory(model)
}

// Update обновляет категорию
func (r *CategoryRepository) Update(ctx context.Context, id int, category *models.EventCategory, categoryType string) (*models.EventCategory, error) {
	strategy, ok := r.strategies[categoryType]
	if !ok {
		return nil, fmt.Errorf("неизвестный тип категории: %s", categoryType)
	}

	// Проверяем существование категории
	model := strategy.GetModel()
	if err := r.db.WithContext(ctx).Table(strategy.GetTableName()).First(model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // возвращаем nil, nil если категория не найдена
		}
		return nil, err
	}

	// Преобразуем и обновляем
	updateModel, err := strategy.ConvertFromEventCategory(category)
	if err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Table(strategy.GetTableName()).Where("id = ?", id).Updates(updateModel).Error; err != nil {
		return nil, err
	}

	// Получаем обновленный объект
	if err := r.db.WithContext(ctx).Table(strategy.GetTableName()).First(model, id).Error; err != nil {
		return nil, err
	}

	return strategy.ConvertToEventCategory(model)
}

// Delete удаляет категорию
func (r *CategoryRepository) Delete(ctx context.Context, id int, categoryType string) error {
	strategy, ok := r.strategies[categoryType]
	if !ok {
		return fmt.Errorf("неизвестный тип категории: %s", categoryType)
	}

	model := strategy.GetModel()
	result := r.db.WithContext(ctx).Table(strategy.GetTableName()).Delete(model, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("категория с id=%d не найдена", id)
	}

	return nil
}

// GetCategoryTypes возвращает список доступных типов категорий
func (r *CategoryRepository) GetCategoryTypes(ctx context.Context) ([]string, error) {
	types := make([]string, 0, len(r.strategies))
	for t := range r.strategies {
		types = append(types, t)
	}
	return types, nil
}

// EventCategoryStrategy стратегия для работы с категориями мероприятий
type EventCategoryStrategy struct{}

func (s *EventCategoryStrategy) GetModel() interface{} {
	return &models.EventCategory{}
}

func (s *EventCategoryStrategy) GetTableName() string {
	return "event_categories"
}

func (s *EventCategoryStrategy) ConvertToEventCategory(model interface{}) (*models.EventCategory, error) {
	category, ok := model.(*models.EventCategory)
	if !ok {
		return nil, fmt.Errorf("неверный тип модели для конвертации")
	}
	return category, nil
}

func (s *EventCategoryStrategy) ConvertFromEventCategory(category *models.EventCategory) (interface{}, error) {
	// Для EventCategory не требуется конвертация
	return category, nil
}

// TransactionCategoryStrategy стратегия для работы с категориями транзакций
type TransactionCategoryStrategy struct{}

func (s *TransactionCategoryStrategy) GetModel() interface{} {
	return &models.TransactionCategory{}
}

func (s *TransactionCategoryStrategy) GetTableName() string {
	return "transaction_categories"
}

func (s *TransactionCategoryStrategy) ConvertToEventCategory(model interface{}) (*models.EventCategory, error) {
	txCategory, ok := model.(*models.TransactionCategory)
	if !ok {
		return nil, fmt.Errorf("неверный тип модели для конвертации")
	}

	// Конвертируем TransactionCategory в EventCategory
	return &models.EventCategory{
		ID:     txCategory.ID,
		Name:   txCategory.Name,
		IconID: txCategory.IconID,
	}, nil
}

func (s *TransactionCategoryStrategy) ConvertFromEventCategory(category *models.EventCategory) (interface{}, error) {
	// Конвертируем EventCategory в TransactionCategory
	return &models.TransactionCategory{
		ID:     category.ID,
		Name:   category.Name,
		IconID: category.IconID,
	}, nil
}

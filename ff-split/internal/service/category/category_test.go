package category

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/mock"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestCategoryService_GetCategories(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockCategory(ctrl)
	categoryService := NewCategoryService(mockRepo)

	ctx := context.Background()
	categoryType := "event"

	t.Run("успешное получение категорий", func(t *testing.T) {
		expectedCategories := []service.CategoryDTO{
			{ID: 1, Name: "Категория 1", IconID: 1},
			{ID: 2, Name: "Категория 2", IconID: 2},
		}

		mockRepo.EXPECT().
			GetAll(ctx, categoryType).
			Return(expectedCategories, nil).
			Times(1)

		result, err := categoryService.GetCategories(ctx, categoryType)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "Категория 1", result[0].Name)
	})

	t.Run("ошибка получения категорий", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockRepo.EXPECT().
			GetAll(ctx, categoryType).
			Return(nil, expectedErr).
			Times(1)

		result, err := categoryService.GetCategories(ctx, categoryType)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestCategoryService_GetCategoryByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockCategory(ctrl)
	categoryService := NewCategoryService(mockRepo)

	ctx := context.Background()
	categoryID := 1
	categoryType := "event"

	t.Run("успешное получение категории", func(t *testing.T) {
		expectedCategory := &service.CategoryDTO{
			ID:     categoryID,
			Name:   "Категория 1",
			IconID: 1,
		}

		mockRepo.EXPECT().
			GetByID(ctx, categoryType, categoryID).
			Return(expectedCategory, nil).
			Times(1)

		result, err := categoryService.GetCategoryByID(ctx, categoryID, categoryType)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, categoryID, result.ID)
		assert.Equal(t, "Категория 1", result.Name)
	})

	t.Run("категория не найдена", func(t *testing.T) {
		expectedErr := errors.New("category not found")
		mockRepo.EXPECT().
			GetByID(ctx, categoryType, categoryID).
			Return(nil, expectedErr).
			Times(1)

		result, err := categoryService.GetCategoryByID(ctx, categoryID, categoryType)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestCategoryService_CreateCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockCategory(ctrl)
	categoryService := NewCategoryService(mockRepo)

	ctx := context.Background()
	categoryType := "event"
	category := &service.CategoryDTO{
		Name:   "Новая категория",
		IconID: 1,
	}

	t.Run("успешное создание категории", func(t *testing.T) {
		expectedCategory := &service.CategoryDTO{
			ID:     1,
			Name:   "Новая категория",
			IconID: 1,
		}

		mockRepo.EXPECT().
			Create(ctx, categoryType, category).
			Return(expectedCategory, nil).
			Times(1)

		result, err := categoryService.CreateCategory(ctx, category, categoryType)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.ID)
		assert.Equal(t, "Новая категория", result.Name)
	})

	t.Run("ошибка создания категории", func(t *testing.T) {
		expectedErr := errors.New("creation error")
		mockRepo.EXPECT().
			Create(ctx, categoryType, category).
			Return(nil, expectedErr).
			Times(1)

		result, err := categoryService.CreateCategory(ctx, category, categoryType)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestCategoryService_UpdateCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockCategory(ctrl)
	categoryService := NewCategoryService(mockRepo)

	ctx := context.Background()
	categoryID := 1
	categoryType := "event"
	category := &service.CategoryDTO{
		ID:     categoryID,
		Name:   "Обновленная категория",
		IconID: 2,
	}

	t.Run("успешное обновление категории", func(t *testing.T) {
		expectedCategory := &service.CategoryDTO{
			ID:     categoryID,
			Name:   "Обновленная категория",
			IconID: 2,
		}

		mockRepo.EXPECT().
			Update(ctx, categoryType, category).
			Return(expectedCategory, nil).
			Times(1)

		result, err := categoryService.UpdateCategory(ctx, categoryID, category, categoryType)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, categoryID, result.ID)
		assert.Equal(t, "Обновленная категория", result.Name)
	})

	t.Run("ошибка обновления категории", func(t *testing.T) {
		expectedErr := errors.New("update error")
		mockRepo.EXPECT().
			Update(ctx, categoryType, category).
			Return(nil, expectedErr).
			Times(1)

		result, err := categoryService.UpdateCategory(ctx, categoryID, category, categoryType)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestCategoryService_DeleteCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockCategory(ctrl)
	categoryService := NewCategoryService(mockRepo)

	ctx := context.Background()
	categoryID := 1
	categoryType := "event"

	t.Run("успешное удаление категории", func(t *testing.T) {
		mockRepo.EXPECT().
			Delete(ctx, categoryType, categoryID).
			Return(nil).
			Times(1)

		err := categoryService.DeleteCategory(ctx, categoryID, categoryType)

		assert.NoError(t, err)
	})

	t.Run("ошибка удаления категории", func(t *testing.T) {
		expectedErr := errors.New("delete error")
		mockRepo.EXPECT().
			Delete(ctx, categoryType, categoryID).
			Return(expectedErr).
			Times(1)

		err := categoryService.DeleteCategory(ctx, categoryID, categoryType)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestCategoryService_GetCategoryTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockCategory(ctrl)
	categoryService := NewCategoryService(mockRepo)

	t.Run("успешное получение типов категорий", func(t *testing.T) {
		expectedTypes := []string{"event", "transaction"}

		mockRepo.EXPECT().
			GetCategoryTypes().
			Return(expectedTypes, nil).
			Times(1)

		result, err := categoryService.GetCategoryTypes()

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
		assert.Contains(t, result, "event")
		assert.Contains(t, result, "transaction")
	})

	t.Run("ошибка получения типов категорий", func(t *testing.T) {
		expectedErr := errors.New("error getting types")
		mockRepo.EXPECT().
			GetCategoryTypes().
			Return(nil, expectedErr).
			Times(1)

		result, err := categoryService.GetCategoryTypes()

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

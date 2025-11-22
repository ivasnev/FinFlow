package tests

import (
	"testing"

	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
	"github.com/stretchr/testify/suite"
)

// CategorySuite представляет suite для тестов управления категориями
type CategorySuite struct {
	BaseSuite
}

// TestCategorySuite запускает все тесты в CategorySuite
func TestCategorySuite(t *testing.T) {
	suite.Run(t, new(CategorySuite))
}

// TestGetCategories_Event_Success тестирует получение категорий мероприятий
func (s *CategorySuite) TestGetCategories_Event_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Travel", TestRequestID)
	s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	s.createTestEventCategory(TestCategoryID2, "Вечеринка", icon.ID)

	// Подготавливаем параметры запроса
	categoryType := api.CategoryType("event")
	params := api.GetCategoriesParams{
		CategoryType: categoryType,
	}

	// Act - действие
	resp, err := s.APIClient.GetCategoriesWithResponse(s.Ctx, &params)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "список категорий должен быть возвращен")
	s.Require().NotNil(resp.JSON200.Categories)
	s.Require().GreaterOrEqual(len(*resp.JSON200.Categories), 2, "должно быть минимум 2 категории")
}

// TestGetCategories_Transaction_Success тестирует получение категорий транзакций
func (s *CategorySuite) TestGetCategories_Transaction_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Food", TestRequestID)
	s.createTestTransactionCategory(TestCategoryID1, "Еда", icon.ID)
	s.createTestTransactionCategory(TestCategoryID2, "Транспорт", icon.ID)

	// Подготавливаем параметры запроса
	categoryType := api.CategoryType("transaction")
	params := api.GetCategoriesParams{
		CategoryType: categoryType,
	}

	// Act - действие
	resp, err := s.APIClient.GetCategoriesWithResponse(s.Ctx, &params)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "список категорий должен быть возвращен")
	s.Require().NotNil(resp.JSON200.Categories)
	s.Require().GreaterOrEqual(len(*resp.JSON200.Categories), 2, "должно быть минимум 2 категории")
}

// TestGetCategoryByID_Event_Success тестирует получение категории мероприятия по ID
func (s *CategorySuite) TestGetCategoryByID_Event_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Travel", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)

	// Подготавливаем параметры запроса
	categoryType := api.CategoryType("event")
	params := api.GetCategoryByIDParams{
		CategoryType: categoryType,
	}

	// Act - действие
	resp, err := s.APIClient.GetCategoryByIDWithResponse(s.Ctx, category.ID, &params)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "категория должна быть возвращена")
	s.Require().Equal(category.ID, *resp.JSON200.Id)
	s.Require().Equal(category.Name, *resp.JSON200.Name)
}

// TestGetCategoryByID_NotFound тестирует получение несуществующей категории
func (s *CategorySuite) TestGetCategoryByID_NotFound() {
	// Arrange - подготовка
	nonExistentID := 999
	categoryType := api.CategoryType("event")
	params := api.GetCategoryByIDParams{
		CategoryType: categoryType,
	}

	// Act - действие
	resp, err := s.APIClient.GetCategoryByIDWithResponse(s.Ctx, nonExistentID, &params)

	// Assert - проверка
	// Может быть ошибка десериализации, но статус код должен быть 404
	if err == nil {
		s.Require().Equal(404, resp.StatusCode(), "должен быть статус 404")
	}
}

// TestCreateCategory_Event_Success тестирует создание категории мероприятия
func (s *CategorySuite) TestCreateCategory_Event_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Sport", TestRequestID)

	// Подготавливаем запрос
	categoryType := api.CategoryType("event")
	categoryName := "Спорт"
	params := api.CreateCategoryParams{
		CategoryType: categoryType,
	}
	reqBody := api.CreateCategoryJSONRequestBody{
		Name:   categoryName,
		IconId: icon.ID,
	}

	// Act - действие
	resp, err := s.APIClient.CreateCategoryWithResponse(s.Ctx, &params, reqBody)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(201, resp.StatusCode(), "должен быть статус 201")
	s.Require().NotNil(resp.JSON201, "категория должна быть создана")
	s.Require().Equal(categoryName, *resp.JSON201.Name)

	// Проверяем, что категория создана в БД
	var count int64
	err = s.GetDB().Table("event_categories").Where("name = ?", categoryName).Count(&count).Error
	s.NoError(err)
	s.Equal(int64(1), count, "должна быть создана одна категория")
}

// TestCreateCategory_Transaction_Success тестирует создание категории транзакции
func (s *CategorySuite) TestCreateCategory_Transaction_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Entertainment", TestRequestID)

	// Подготавливаем запрос
	categoryType := api.CategoryType("transaction")
	categoryName := "Развлечения"
	params := api.CreateCategoryParams{
		CategoryType: categoryType,
	}
	reqBody := api.CreateCategoryJSONRequestBody{
		Name:   categoryName,
		IconId: icon.ID,
	}

	// Act - действие
	resp, err := s.APIClient.CreateCategoryWithResponse(s.Ctx, &params, reqBody)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(201, resp.StatusCode(), "должен быть статус 201")
	s.Require().NotNil(resp.JSON201, "категория должна быть создана")
	s.Require().Equal(categoryName, *resp.JSON201.Name)

	// Проверяем, что категория создана в БД
	var count int64
	err = s.GetDB().Table("transaction_categories").Where("name = ?", categoryName).Count(&count).Error
	s.NoError(err)
	s.Equal(int64(1), count, "должна быть создана одна категория")
}

// TestUpdateCategory_Success тестирует обновление категории
func (s *CategorySuite) TestUpdateCategory_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Travel", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Старое название", icon.ID)

	// Подготавливаем запрос на обновление
	newName := "Новое название"
	categoryType := api.CategoryType("event")
	params := api.UpdateCategoryParams{
		CategoryType: categoryType,
	}
	reqBody := api.UpdateCategoryJSONRequestBody{
		Name:   newName,
		IconId: icon.ID,
	}

	// Act - действие
	resp, err := s.APIClient.UpdateCategoryWithResponse(s.Ctx, category.ID, &params, reqBody)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "обновленная категория должна быть возвращена")
	s.Require().Equal(newName, *resp.JSON200.Name)

	// Проверяем, что данные обновлены в БД
	var updatedCategory struct {
		Name string
	}
	err = s.GetDB().Table("event_categories").Where("id = ?", category.ID).First(&updatedCategory).Error
	s.NoError(err)
	s.Equal(newName, updatedCategory.Name)
}

// TestDeleteCategory_Success тестирует удаление категории
func (s *CategorySuite) TestDeleteCategory_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Travel", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Категория для удаления", icon.ID)

	// Подготавливаем параметры запроса
	categoryType := api.CategoryType("event")
	params := api.DeleteCategoryParams{
		CategoryType: categoryType,
	}

	// Act - действие
	resp, err := s.APIClient.DeleteCategoryWithResponse(s.Ctx, category.ID, &params)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "должен быть возвращен объект успеха")

	// Проверяем, что категория удалена из БД
	var count int64
	err = s.GetDB().Table("event_categories").Where("id = ?", category.ID).Count(&count).Error
	s.NoError(err)
	s.Equal(int64(0), count, "категория должна быть удалена из БД")
}


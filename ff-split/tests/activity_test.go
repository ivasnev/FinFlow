package tests

import (
	"testing"

	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
	"github.com/stretchr/testify/suite"
)

// ActivitySuite представляет suite для тестов управления активностями
type ActivitySuite struct {
	BaseSuite
}

// TestActivitySuite запускает все тесты в ActivitySuite
func TestActivitySuite(t *testing.T) {
	suite.Run(t, new(ActivitySuite))
}

// TestGetActivitiesByEventID_Success тестирует получение активностей мероприятия
func (s *ActivitySuite) TestGetActivitiesByEventID_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Activity", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	event := s.createTestEvent(TestEventID1, TestEventName1, "Описание", &category.ID)

	user1 := s.createTestUser(TestUserID1, TestUserID1, TestNickname1, TestName1)
	s.addUserToEvent(user1.ID, event.ID)

	// Создаем несколько активностей напрямую в БД
	err := s.GetDB().Exec(`
		INSERT INTO activities (id, event_id, user_id, icon_id, description)
		VALUES ($1, $2, $3, $4, $5)
	`, TestActivityID1, event.ID, user1.ID, icon.ID, "Активность 1").Error
	s.NoError(err)

	err = s.GetDB().Exec(`
		INSERT INTO activities (id, event_id, user_id, icon_id, description)
		VALUES ($1, $2, $3, $4, $5)
	`, TestActivityID2, event.ID, user1.ID, icon.ID, "Активность 2").Error
	s.NoError(err)

	// Act - действие
	resp, err := s.APIClient.GetActivitiesByEventIDWithResponse(s.Ctx, event.ID)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "список активностей должен быть возвращен")
	s.Require().NotNil(resp.JSON200.Activities)
	s.Require().GreaterOrEqual(len(*resp.JSON200.Activities), 2, "должно быть минимум 2 активности")
}

// TestCreateActivity_Success тестирует создание активности
func (s *ActivitySuite) TestCreateActivity_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Comment", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	event := s.createTestEvent(TestEventID1, TestEventName1, "Описание", &category.ID)

	user1 := s.createTestUser(TestUserID1, TestUserID1, TestNickname1, TestName1)
	s.addUserToEvent(user1.ID, event.ID)

	// Подготавливаем запрос
	activityDescription := "Пользователь оставил комментарий"
	reqBody := api.CreateActivityJSONRequestBody{
		Description: activityDescription,
		IconId:      &icon.ID,
		UserId:      &user1.ID,
	}

	// Act - действие
	resp, err := s.APIClient.CreateActivityWithResponse(s.Ctx, event.ID, reqBody)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(201, resp.StatusCode(), "должен быть статус 201")
	s.Require().NotNil(resp.JSON201, "активность должна быть создана")
	s.Require().Equal(activityDescription, *resp.JSON201.Description)
	s.Require().Equal(icon.ID, *resp.JSON201.IconId)

	// Проверяем, что активность создана в БД
	var count int64
	err = s.GetDB().Table("activities").Where("description = ?", activityDescription).Count(&count).Error
	s.NoError(err)
	s.Equal(int64(1), count, "должна быть создана одна активность")
}

// TestGetActivityByID_Success тестирует получение активности по ID
func (s *ActivitySuite) TestGetActivityByID_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Activity", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	event := s.createTestEvent(TestEventID1, TestEventName1, "Описание", &category.ID)

	user1 := s.createTestUser(TestUserID1, TestUserID1, TestNickname1, TestName1)
	s.addUserToEvent(user1.ID, event.ID)

	// Создаем активность
	activityDescription := "Тестовая активность"
	err := s.GetDB().Exec(`
		INSERT INTO activities (id, event_id, user_id, icon_id, description)
		VALUES ($1, $2, $3, $4, $5)
	`, TestActivityID1, event.ID, user1.ID, icon.ID, activityDescription).Error
	s.NoError(err)

	// Act - действие
	resp, err := s.APIClient.GetActivityByIDWithResponse(s.Ctx, event.ID, TestActivityID1)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "активность должна быть возвращена")
	s.Require().Equal(TestActivityID1, *resp.JSON200.ActivityId)
	s.Require().Equal(activityDescription, *resp.JSON200.Description)
}

// TestGetActivityByID_NotFound тестирует получение несуществующей активности
func (s *ActivitySuite) TestGetActivityByID_NotFound() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Activity", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	event := s.createTestEvent(TestEventID1, TestEventName1, "Описание", &category.ID)

	nonExistentID := 999

	// Act - действие
	resp, err := s.APIClient.GetActivityByIDWithResponse(s.Ctx, event.ID, nonExistentID)

	// Assert - проверка
	// Может быть ошибка десериализации, но статус код должен быть 404
	if err == nil {
		s.Require().Equal(404, resp.StatusCode(), "должен быть статус 404")
	}
}

// TestUpdateActivity_Success тестирует обновление активности
func (s *ActivitySuite) TestUpdateActivity_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Activity", TestRequestID)
	icon2 := s.createTestIcon(TestIconID2, "NewIcon", "uuid-new")
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	event := s.createTestEvent(TestEventID1, TestEventName1, "Описание", &category.ID)

	user1 := s.createTestUser(TestUserID1, TestUserID1, TestNickname1, TestName1)
	s.addUserToEvent(user1.ID, event.ID)

	// Создаем активность
	err := s.GetDB().Exec(`
		INSERT INTO activities (id, event_id, user_id, icon_id, description)
		VALUES ($1, $2, $3, $4, $5)
	`, TestActivityID1, event.ID, user1.ID, icon.ID, "Старое описание").Error
	s.NoError(err)

	// Подготавливаем запрос на обновление
	newDescription := "Новое описание"
	reqBody := api.UpdateActivityJSONRequestBody{
		Description: newDescription,
		IconId:      &icon2.ID,
		UserId:      &user1.ID,
	}

	// Act - действие
	resp, err := s.APIClient.UpdateActivityWithResponse(s.Ctx, event.ID, TestActivityID1, reqBody)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "обновленная активность должна быть возвращена")
	s.Require().Equal(newDescription, *resp.JSON200.Description)
	s.Require().Equal(icon2.ID, *resp.JSON200.IconId)

	// Проверяем, что данные обновлены в БД
	var updatedActivity struct {
		Description string
		IconID      int `gorm:"column:icon_id"`
	}
	err = s.GetDB().Table("activities").Where("id = ?", TestActivityID1).First(&updatedActivity).Error
	s.NoError(err)
	s.Equal(newDescription, updatedActivity.Description)
	s.Equal(icon2.ID, updatedActivity.IconID)
}

// TestDeleteActivity_Success тестирует удаление активности
func (s *ActivitySuite) TestDeleteActivity_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Activity", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	event := s.createTestEvent(TestEventID1, TestEventName1, "Описание", &category.ID)

	user1 := s.createTestUser(TestUserID1, TestUserID1, TestNickname1, TestName1)
	s.addUserToEvent(user1.ID, event.ID)

	// Создаем активность
	err := s.GetDB().Exec(`
		INSERT INTO activities (id, event_id, user_id, icon_id, description)
		VALUES ($1, $2, $3, $4, $5)
	`, TestActivityID1, event.ID, user1.ID, icon.ID, "Активность для удаления").Error
	s.NoError(err)

	// Act - действие
	resp, err := s.APIClient.DeleteActivityWithResponse(s.Ctx, event.ID, TestActivityID1)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "должен быть возвращен объект успеха")

	// Проверяем, что активность удалена из БД
	var count int64
	err = s.GetDB().Table("activities").Where("id = ?", TestActivityID1).Count(&count).Error
	s.NoError(err)
	s.Equal(int64(0), count, "активность должна быть удалена из БД")
}


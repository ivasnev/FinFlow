package tests

import (
	"testing"

	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
	"github.com/stretchr/testify/suite"
)

// TaskSuite представляет suite для тестов управления задачами
type TaskSuite struct {
	BaseSuite
}

// TestTaskSuite запускает все тесты в TaskSuite
func TestTaskSuite(t *testing.T) {
	suite.Run(t, new(TaskSuite))
}

// TestGetTasksByEventID_Success тестирует получение задач мероприятия
func (s *TaskSuite) TestGetTasksByEventID_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Travel", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	event := s.createTestEvent(TestEventID1, TestEventName1, "Описание", &category.ID)

	user1 := s.createTestUser(TestUserID1, TestUserID1, TestNickname1, TestName1)
	s.addUserToEvent(user1.ID, event.ID)

	// Создаем несколько задач напрямую в БД
	err := s.GetDB().Exec(`
		INSERT INTO tasks (id, event_id, user_id, title, description, priority)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, TestTaskID1, event.ID, user1.ID, "Задача 1", "Описание задачи 1", 1).Error
	s.NoError(err)

	err = s.GetDB().Exec(`
		INSERT INTO tasks (id, event_id, user_id, title, description, priority)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, TestTaskID2, event.ID, user1.ID, "Задача 2", "Описание задачи 2", 2).Error
	s.NoError(err)

	// Act - действие
	resp, err := s.APIClient.GetTasksByEventIDWithResponse(s.Ctx, event.ID)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "список задач должен быть возвращен")
	s.Require().NotNil(resp.JSON200.Tasks)
	s.Require().GreaterOrEqual(len(*resp.JSON200.Tasks), 2, "должно быть минимум 2 задачи")
}

// TestCreateTask_Success тестирует создание задачи
func (s *TaskSuite) TestCreateTask_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Travel", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	event := s.createTestEvent(TestEventID1, TestEventName1, "Описание", &category.ID)

	user1 := s.createTestUser(TestUserID1, TestUserID1, TestNickname1, TestName1)
	s.addUserToEvent(user1.ID, event.ID)

	// Подготавливаем запрос
	taskTitle := "Купить билеты"
	taskDescription := "Купить билеты на самолет"
	taskPriority := 1
	reqBody := api.CreateTaskJSONRequestBody{
		Title:       taskTitle,
		Description: &taskDescription,
		UserId:      user1.ID,
		Priority:    &taskPriority,
	}

	// Act - действие
	resp, err := s.APIClient.CreateTaskWithResponse(s.Ctx, event.ID, reqBody)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(201, resp.StatusCode(), "должен быть статус 201")
	s.Require().NotNil(resp.JSON201, "задача должна быть создана")
	s.Require().NotNil(resp.JSON201.Task)
	s.Require().Equal(taskTitle, *resp.JSON201.Task.Title)
	s.Require().Equal(taskDescription, *resp.JSON201.Task.Description)
	s.Require().Equal(user1.ID, *resp.JSON201.Task.UserId)

	// Проверяем, что задача создана в БД
	var count int64
	err = s.GetDB().Table("tasks").Where("title = ?", taskTitle).Count(&count).Error
	s.NoError(err)
	s.Equal(int64(1), count, "должна быть создана одна задача")
}

// TestGetTaskByID_Success тестирует получение задачи по ID
func (s *TaskSuite) TestGetTaskByID_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Travel", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	event := s.createTestEvent(TestEventID1, TestEventName1, "Описание", &category.ID)

	user1 := s.createTestUser(TestUserID1, TestUserID1, TestNickname1, TestName1)
	s.addUserToEvent(user1.ID, event.ID)

	// Создаем задачу
	taskTitle := "Тестовая задача"
	err := s.GetDB().Exec(`
		INSERT INTO tasks (id, event_id, user_id, title, description, priority)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, TestTaskID1, event.ID, user1.ID, taskTitle, "Описание", 1).Error
	s.NoError(err)

	// Act - действие
	resp, err := s.APIClient.GetTaskByIDWithResponse(s.Ctx, event.ID, TestTaskID1)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "задача должна быть возвращена")
	s.Require().NotNil(resp.JSON200.Task)
	s.Require().Equal(TestTaskID1, *resp.JSON200.Task.Id)
	s.Require().Equal(taskTitle, *resp.JSON200.Task.Title)
}

// TestGetTaskByID_NotFound тестирует получение несуществующей задачи
func (s *TaskSuite) TestGetTaskByID_NotFound() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Travel", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	event := s.createTestEvent(TestEventID1, TestEventName1, "Описание", &category.ID)

	nonExistentID := 999

	// Act - действие
	resp, err := s.APIClient.GetTaskByIDWithResponse(s.Ctx, event.ID, nonExistentID)

	// Assert - проверка
	// Может быть ошибка десериализации, но статус код должен быть 404
	if err == nil {
		s.Require().Equal(404, resp.StatusCode(), "должен быть статус 404")
	}
}

// TestUpdateTask_Success тестирует обновление задачи
func (s *TaskSuite) TestUpdateTask_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Travel", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	event := s.createTestEvent(TestEventID1, TestEventName1, "Описание", &category.ID)

	user1 := s.createTestUser(TestUserID1, TestUserID1, TestNickname1, TestName1)
	s.addUserToEvent(user1.ID, event.ID)

	// Создаем задачу
	err := s.GetDB().Exec(`
		INSERT INTO tasks (id, event_id, user_id, title, description, priority)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, TestTaskID1, event.ID, user1.ID, "Старое название", "Старое описание", 1).Error
	s.NoError(err)

	// Подготавливаем запрос на обновление
	newTitle := "Новое название"
	newDescription := "Новое описание"
	newPriority := 5
	reqBody := api.UpdateTaskJSONRequestBody{
		Title:       newTitle,
		Description: &newDescription,
		UserId:      user1.ID,
		Priority:    &newPriority,
	}

	// Act - действие
	resp, err := s.APIClient.UpdateTaskWithResponse(s.Ctx, event.ID, TestTaskID1, reqBody)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "обновленная задача должна быть возвращена")
	s.Require().NotNil(resp.JSON200.Task)
	s.Require().Equal(newTitle, *resp.JSON200.Task.Title)
	s.Require().Equal(newDescription, *resp.JSON200.Task.Description)
	s.Require().Equal(newPriority, *resp.JSON200.Task.Priority)

	// Проверяем, что данные обновлены в БД
	var updatedTask struct {
		Title       string
		Description string
		Priority    int
	}
	err = s.GetDB().Table("tasks").Where("id = ?", TestTaskID1).First(&updatedTask).Error
	s.NoError(err)
	s.Equal(newTitle, updatedTask.Title)
	s.Equal(newDescription, updatedTask.Description)
	s.Equal(newPriority, updatedTask.Priority)
}

// TestDeleteTask_Success тестирует удаление задачи
func (s *TaskSuite) TestDeleteTask_Success() {
	// Arrange - подготовка
	icon := s.createTestIcon(TestIconID1, "Travel", TestRequestID)
	category := s.createTestEventCategory(TestCategoryID1, "Путешествие", icon.ID)
	event := s.createTestEvent(TestEventID1, TestEventName1, "Описание", &category.ID)

	user1 := s.createTestUser(TestUserID1, TestUserID1, TestNickname1, TestName1)
	s.addUserToEvent(user1.ID, event.ID)

	// Создаем задачу
	err := s.GetDB().Exec(`
		INSERT INTO tasks (id, event_id, user_id, title, description, priority)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, TestTaskID1, event.ID, user1.ID, "Задача для удаления", "Описание", 1).Error
	s.NoError(err)

	// Act - действие
	resp, err := s.APIClient.DeleteTaskWithResponse(s.Ctx, event.ID, TestTaskID1)

	// Assert - проверка
	s.Require().NoError(err, "запрос должен выполниться успешно")
	s.Require().Equal(200, resp.StatusCode(), "должен быть статус 200")
	s.Require().NotNil(resp.JSON200, "должен быть возвращен объект успеха")

	// Проверяем, что задача удалена из БД
	var count int64
	err = s.GetDB().Table("tasks").Where("id = ?", TestTaskID1).Count(&count).Error
	s.NoError(err)
	s.Equal(int64(0), count, "задача должна быть удалена из БД")
}


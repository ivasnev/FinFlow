package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	repositoryMock "github.com/ivasnev/FinFlow/ff-split/internal/repository/mock"
	serviceMock "github.com/ivasnev/FinFlow/ff-split/internal/service/mock"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

func TestTaskService_GetTasksByEventID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := repositoryMock.NewMockTask(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	taskService := NewTaskService(mockTaskRepo, mockUserService)

	eventID := int64(1)

	t.Run("успешное получение задач", func(t *testing.T) {
		userID := int64(100)
		tasks := []models.Task{
			{
				ID:          1,
				UserID:      &userID,
				EventID:     &eventID,
				Title:       "Task 1",
				Description: "Description 1",
				Priority:    1,
				CreatedAt:   time.Now(),
			},
			{
				ID:          2,
				UserID:      &userID,
				EventID:     &eventID,
				Title:       "Task 2",
				Description: "Description 2",
				Priority:    2,
				CreatedAt:   time.Now(),
			},
		}

		mockTaskRepo.EXPECT().
			GetTasksByEventID(eventID).
			Return(tasks, nil).
			Times(1)

		result, err := taskService.GetTasksByEventID(context.Background(), eventID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 2, len(result))
		assert.Equal(t, "Task 1", result[0].Title)
	})

	t.Run("ошибка получения задач", func(t *testing.T) {
		expectedErr := errors.New("database error")
		mockTaskRepo.EXPECT().
			GetTasksByEventID(eventID).
			Return(nil, expectedErr).
			Times(1)

		result, err := taskService.GetTasksByEventID(context.Background(), eventID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTaskService_GetTaskByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := repositoryMock.NewMockTask(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	taskService := NewTaskService(mockTaskRepo, mockUserService)

	taskID := uint(1)

	t.Run("успешное получение задачи", func(t *testing.T) {
		userID := int64(100)
		eventID := int64(1)
		task := &models.Task{
			ID:          1,
			UserID:      &userID,
			EventID:     &eventID,
			Title:       "Test Task",
			Description: "Test Description",
			Priority:    1,
			CreatedAt:   time.Now(),
		}

		mockTaskRepo.EXPECT().
			GetTaskByID(taskID).
			Return(task, nil).
			Times(1)

		result, err := taskService.GetTaskByID(context.Background(), taskID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, taskID, result.ID)
		assert.Equal(t, "Test Task", result.Title)
	})

	t.Run("задача не найдена", func(t *testing.T) {
		expectedErr := errors.New("task not found")
		mockTaskRepo.EXPECT().
			GetTaskByID(taskID).
			Return(nil, expectedErr).
			Times(1)

		result, err := taskService.GetTaskByID(context.Background(), taskID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTaskService_CreateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := repositoryMock.NewMockTask(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	taskService := NewTaskService(mockTaskRepo, mockUserService)

	ctx := context.Background()
	eventID := int64(1)
	userID := int64(100)
	internalUserID := int64(1)

	t.Run("успешное создание задачи", func(t *testing.T) {
		taskRequest := &service.TaskRequest{
			UserID:      userID,
			Title:       "New Task",
			Description: "New Description",
			Priority:    1,
		}

		user := &models.User{
			ID:     internalUserID,
			UserID: &userID,
		}

		mockUserService.EXPECT().
			GetUserByInternalUserID(ctx, userID).
			Return(user, nil).
			Times(1)

		mockTaskRepo.EXPECT().
			CreateTask(gomock.Any()).
			DoAndReturn(func(task *models.Task) error {
				assert.Equal(t, internalUserID, *task.UserID)
				assert.Equal(t, eventID, *task.EventID)
				assert.Equal(t, "New Task", task.Title)
				task.ID = 1
				task.CreatedAt = time.Now()
				return nil
			}).
			Times(1)

		result, err := taskService.CreateTask(ctx, eventID, taskRequest)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "New Task", result.Title)
		assert.Equal(t, internalUserID, result.UserID)
		assert.Equal(t, eventID, result.EventID)
	})

	t.Run("пользователь не найден", func(t *testing.T) {
		taskRequest := &service.TaskRequest{
			UserID: userID,
			Title:  "New Task",
		}

		expectedErr := errors.New("user not found")
		mockUserService.EXPECT().
			GetUserByInternalUserID(ctx, userID).
			Return(nil, expectedErr).
			Times(1)

		result, err := taskService.CreateTask(ctx, eventID, taskRequest)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTaskService_UpdateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := repositoryMock.NewMockTask(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	taskService := NewTaskService(mockTaskRepo, mockUserService)

	ctx := context.Background()
	taskID := uint(1)
	userID := int64(100)
	internalUserID := int64(1)

	t.Run("успешное обновление задачи", func(t *testing.T) {
		taskRequest := &service.TaskRequest{
			UserID:      userID,
			Title:       "Updated Task",
			Description: "Updated Description",
			Priority:    2,
		}

		user := &models.User{
			ID:     internalUserID,
			UserID: &userID,
		}

		existingTask := &models.Task{
			ID:          1,
			UserID:      &internalUserID,
			EventID:     func() *int64 { id := int64(1); return &id }(),
			Title:       "Old Task",
			Description: "Old Description",
			Priority:    1,
			CreatedAt:   time.Now(),
		}

		mockUserService.EXPECT().
			GetUserByInternalUserID(ctx, userID).
			Return(user, nil).
			Times(1)

		mockTaskRepo.EXPECT().
			GetTaskByID(taskID).
			Return(existingTask, nil).
			Times(1)

		mockTaskRepo.EXPECT().
			UpdateTask(gomock.Any()).
			DoAndReturn(func(task *models.Task) error {
				assert.Equal(t, internalUserID, *task.UserID)
				assert.Equal(t, "Updated Task", task.Title)
				assert.Equal(t, "Updated Description", task.Description)
				assert.Equal(t, 2, task.Priority)
				return nil
			}).
			Times(1)

		result, err := taskService.UpdateTask(ctx, taskID, taskRequest)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Updated Task", result.Title)
	})

	t.Run("задача не найдена", func(t *testing.T) {
		taskRequest := &service.TaskRequest{
			UserID: userID,
			Title:  "Updated Task",
		}

		user := &models.User{
			ID:     internalUserID,
			UserID: &userID,
		}

		mockUserService.EXPECT().
			GetUserByInternalUserID(ctx, userID).
			Return(user, nil).
			Times(1)

		expectedErr := errors.New("task not found")
		mockTaskRepo.EXPECT().
			GetTaskByID(taskID).
			Return(nil, expectedErr).
			Times(1)

		result, err := taskService.UpdateTask(ctx, taskID, taskRequest)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, expectedErr)
	})
}

func TestTaskService_DeleteTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTaskRepo := repositoryMock.NewMockTask(ctrl)
	mockUserService := serviceMock.NewMockUser(ctrl)
	taskService := NewTaskService(mockTaskRepo, mockUserService)

	taskID := uint(1)

	t.Run("успешное удаление задачи", func(t *testing.T) {
		mockTaskRepo.EXPECT().
			DeleteTask(taskID).
			Return(nil).
			Times(1)

		err := taskService.DeleteTask(context.Background(), taskID)

		assert.NoError(t, err)
	})

	t.Run("ошибка удаления задачи", func(t *testing.T) {
		expectedErr := errors.New("delete error")
		mockTaskRepo.EXPECT().
			DeleteTask(taskID).
			Return(expectedErr).
			Times(1)

		err := taskService.DeleteTask(context.Background(), taskID)

		assert.Error(t, err)
		assert.ErrorIs(t, err, expectedErr)
	})
}


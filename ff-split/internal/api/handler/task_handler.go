package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
	"github.com/ivasnev/FinFlow/ff-split/pkg/api"
)

// GetTasksByEventID возвращает задачи мероприятия
func (s *ServerHandler) GetTasksByEventID(c *gin.Context, idEvent int64) {
	tasks, err := s.taskService.GetTasksByEventID(c.Request.Context(), idEvent)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении задач: %w", err))
		return
	}

	apiTasks := make([]api.TaskDTO, 0, len(tasks))
	for _, t := range tasks {
		apiTasks = append(apiTasks, convertTaskToAPI(&t))
	}

	c.JSON(http.StatusOK, api.TaskListResponse{Tasks: &apiTasks})
}

// GetTaskByID возвращает задачу по ID
func (s *ServerHandler) GetTaskByID(c *gin.Context, idEvent int64, idTask int) {
	task, err := s.taskService.GetTaskByID(c.Request.Context(), uint(idTask))
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при получении задачи: %w", err))
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, api.ErrorResponse{Error: "задача не найдена"})
		return
	}

	c.JSON(http.StatusOK, api.TaskResponse{Task: convertTaskToAPIPtr(task)})
}

// CreateTask создает новую задачу
func (s *ServerHandler) CreateTask(c *gin.Context, idEvent int64) {
	var apiRequest api.TaskRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "некорректные данные запроса"})
		return
	}

	dtoRequest := service.TaskRequest{
		UserID: apiRequest.UserId,
		Title:  apiRequest.Title,
	}

	if apiRequest.Description != nil {
		dtoRequest.Description = *apiRequest.Description
	}

	if apiRequest.Priority != nil {
		dtoRequest.Priority = *apiRequest.Priority
	}

	task, err := s.taskService.CreateTask(c.Request.Context(), idEvent, &dtoRequest)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при создании задачи: %w", err))
		return
	}

	c.JSON(http.StatusCreated, api.TaskResponse{Task: convertTaskToAPIPtr(task)})
}

// UpdateTask обновляет задачу
func (s *ServerHandler) UpdateTask(c *gin.Context, idEvent int64, idTask int) {
	var apiRequest api.TaskRequest
	if err := c.ShouldBindJSON(&apiRequest); err != nil {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{Error: "некорректные данные запроса"})
		return
	}

	dtoRequest := service.TaskRequest{
		UserID: apiRequest.UserId,
		Title:  apiRequest.Title,
	}

	if apiRequest.Description != nil {
		dtoRequest.Description = *apiRequest.Description
	}

	if apiRequest.Priority != nil {
		dtoRequest.Priority = *apiRequest.Priority
	}

	task, err := s.taskService.UpdateTask(c.Request.Context(), uint(idTask), &dtoRequest)
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при обновлении задачи: %w", err))
		return
	}

	c.JSON(http.StatusOK, api.TaskResponse{Task: convertTaskToAPIPtr(task)})
}

// DeleteTask удаляет задачу
func (s *ServerHandler) DeleteTask(c *gin.Context, idEvent int64, idTask int) {
	err := s.taskService.DeleteTask(c.Request.Context(), uint(idTask))
	if err != nil {
		errors.HTTPErrorHandler(c, fmt.Errorf("ошибка при удалении задачи: %w", err))
		return
	}

	c.JSON(http.StatusOK, api.SuccessResponse{Success: true})
}

// Helper functions

func convertTaskToAPI(t *service.TaskDTO) api.TaskDTO {
	id := int(t.ID)
	return api.TaskDTO{
		Id:          &id,
		EventId:     &t.EventID,
		UserId:      &t.UserID,
		Title:       &t.Title,
		Description: &t.Description,
		Priority:    &t.Priority,
		CreatedAt:   &t.CreatedAt,
	}
}

func convertTaskToAPIPtr(t *service.TaskDTO) *api.TaskDTO {
	task := convertTaskToAPI(t)
	return &task
}

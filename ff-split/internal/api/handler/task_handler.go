package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// TaskHandler обработчик для работы с задачами
type TaskHandler struct {
	service service.TaskServiceInterface
}

// NewTaskHandler создает новый обработчик для работы с задачами
func NewTaskHandler(service service.TaskServiceInterface) *TaskHandler {
	return &TaskHandler{service: service}
}

// GetTasksByEventID возвращает список всех задач мероприятия
func (h *TaskHandler) GetTasksByEventID(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID мероприятия"})
		return
	}

	tasks, err := h.service.GetTasksByEventID(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.TaskListResponse{Tasks: tasks})
}

// GetTaskByID возвращает задачу по ID
func (h *TaskHandler) GetTaskByID(c *gin.Context) {
	idStr := c.Param("id_task")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID задачи"})
		return
	}

	task, err := h.service.GetTaskByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.TaskResponse{Task: *task})
}

// CreateTask создает новую задачу
func (h *TaskHandler) CreateTask(c *gin.Context) {
	eventID, err := strconv.ParseInt(c.Param("id_event"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID мероприятия"})
		return
	}

	var taskRequest dto.TaskRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	createdTask, err := h.service.CreateTask(c.Request.Context(), eventID, &taskRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.TaskResponse{Task: *createdTask})
}

// UpdateTask обновляет существующую задачу
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	idStr := c.Param("id_task")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID задачи"})
		return
	}

	var taskRequest dto.TaskRequest
	if err := c.ShouldBindJSON(&taskRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedTask, err := h.service.UpdateTask(c.Request.Context(), uint(id), &taskRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.TaskResponse{Task: *updatedTask})
}

// DeleteTask удаляет задачу по ID
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	idStr := c.Param("id_task")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат ID задачи"})
		return
	}

	err = h.service.DeleteTask(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

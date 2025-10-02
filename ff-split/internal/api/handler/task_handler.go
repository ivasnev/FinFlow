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
// @Summary Получить все задачи мероприятия
// @Description Возвращает список всех задач, связанных с указанным мероприятием
// @Tags задачи
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Success 200 {object} dto.TaskListResponse "Список задач"
// @Failure 400 {object} map[string]string "Неверный формат ID мероприятия"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/task [get]
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
// @Summary Получить задачу по ID
// @Description Возвращает информацию о конкретной задаче по её ID
// @Tags задачи
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id_task path int true "ID задачи"
// @Success 200 {object} dto.TaskResponse "Информация о задаче"
// @Failure 400 {object} map[string]string "Неверный формат ID"
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/task/{id_task} [get]
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
// @Summary Создать новую задачу
// @Description Создает новую задачу в рамках указанного мероприятия
// @Tags задачи
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param task body dto.TaskRequest true "Данные задачи"
// @Success 201 {object} dto.TaskResponse "Созданная задача"
// @Failure 400 {object} map[string]string "Неверный формат данных запроса"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/task [post]
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
// @Summary Обновить задачу
// @Description Обновляет существующую задачу по ID
// @Tags задачи
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id_task path int true "ID задачи"
// @Param task body dto.TaskRequest true "Данные задачи"
// @Success 200 {object} dto.TaskResponse "Обновленная задача"
// @Failure 400 {object} map[string]string "Неверный формат данных запроса"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/task/{id_task} [put]
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
// @Summary Удалить задачу
// @Description Удаляет задачу по ID
// @Tags задачи
// @Accept json
// @Produce json
// @Param id_event path int true "ID мероприятия"
// @Param id_task path int true "ID задачи"
// @Success 204 "Задача успешно удалена"
// @Failure 400 {object} map[string]string "Неверный формат ID"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Router /api/v1/event/{id_event}/task/{id_task} [delete]
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

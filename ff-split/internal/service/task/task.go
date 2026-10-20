package task

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"github.com/ivasnev/FinFlow/ff-split/internal/service"
)

// TaskService реализует сервис для работы с задачами
type TaskService struct {
	repo        repository.Task
	userService service.User
}

// NewTaskService создает новый сервис для работы с задачами
func NewTaskService(repo repository.Task, userService service.User) *TaskService {
	return &TaskService{repo: repo, userService: userService}
}

// GetTasksByEventID возвращает список задач мероприятия
func (s *TaskService) GetTasksByEventID(ctx context.Context, eventID int64) ([]service.TaskDTO, error) {
	tasks, err := s.contextRepo(ctx).GetTasksByEventID(eventID)
	if err != nil {
		return nil, err
	}

	taskDTOs := make([]service.TaskDTO, len(tasks))
	for i, task := range tasks {
		taskDTOs[i] = mapTaskToDTO(task)
	}

	return taskDTOs, nil
}

// GetTaskByID возвращает задачу по ID
func (s *TaskService) GetTaskByID(ctx context.Context, id uint) (*service.TaskDTO, error) {
	task, err := s.contextRepo(ctx).GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	taskDTO := mapTaskToDTO(*task)
	return &taskDTO, nil
}

// CreateTask создает новую задачу
func (s *TaskService) CreateTask(ctx context.Context, eventID int64, taskRequest *service.TaskRequest) (*service.TaskDTO, error) {
	user, err := s.userService.GetUserByInternalUserID(ctx, taskRequest.UserID)
	if err != nil {
		return nil, err
	}

	task := models.Task{
		UserID:      &user.ID,
		EventID:     &eventID,
		Title:       taskRequest.Title,
		Description: taskRequest.Description,
		Priority:    taskRequest.Priority,
	}

	err = s.contextRepo(ctx).CreateTask(&task)
	if err != nil {
		return nil, err
	}

	taskDTO := mapTaskToDTO(task)
	return &taskDTO, nil
}

// UpdateTask обновляет существующую задачу
func (s *TaskService) UpdateTask(ctx context.Context, id uint, taskRequest *service.TaskRequest) (*service.TaskDTO, error) {
	user, err := s.userService.GetUserByInternalUserID(ctx, taskRequest.UserID)
	if err != nil {
		return nil, err
	}

	// Получаем существующую задачу для сохранения неизменяемых полей
	existingTask, err := s.contextRepo(ctx).GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	// Обновляем поля задачи
	existingTask.UserID = &user.ID
	existingTask.Title = taskRequest.Title
	existingTask.Description = taskRequest.Description
	existingTask.Priority = taskRequest.Priority

	err = s.contextRepo(ctx).UpdateTask(existingTask)
	if err != nil {
		return nil, err
	}

	taskDTO := mapTaskToDTO(*existingTask)
	return &taskDTO, nil
}

// DeleteTask удаляет задачу по ID
func (s *TaskService) DeleteTask(ctx context.Context, id uint) error {
	return s.contextRepo(ctx).DeleteTask(id)
}

// Вспомогательные функции для маппинга между моделью и DTO

func mapTaskToDTO(task models.Task) service.TaskDTO {
	var userID, eventID int64
	if task.UserID != nil {
		userID = *task.UserID
	}
	if task.EventID != nil {
		eventID = *task.EventID
	}
	return service.TaskDTO{
		ID:          uint(task.ID),
		UserID:      userID,
		EventID:     eventID,
		Title:       task.Title,
		Description: task.Description,
		Priority:    task.Priority,
		CreatedAt:   task.CreatedAt,
	}
}

// contextRepo передаёт отмену и доверенного автора конкретному репозиторию без изменения общего экземпляра.
func (s *TaskService) contextRepo(ctx context.Context) repository.Task {
	if r, ok := s.repo.(interface {
		WithContext(context.Context) repository.Task
	}); ok {
		return r.WithContext(ctx)
	}
	return s.repo
}

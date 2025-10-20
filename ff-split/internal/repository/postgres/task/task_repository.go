package task

import (
	"errors"

	"github.com/ivasnev/FinFlow/ff-split/internal/models"
	"gorm.io/gorm"
)

// TaskRepository интерфейс репозитория задач
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository создает новый репозиторий задач
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// GetTasksByEventID возвращает список всех задач мероприятия
func (r *TaskRepository) GetTasksByEventID(eventID int64) ([]models.Task, error) {
	var tasks []models.Task
	if err := r.db.Where("event_id = ?", eventID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTaskByID возвращает задачу по идентификатору
func (r *TaskRepository) GetTaskByID(id uint) (*models.Task, error) {
	var task models.Task
	if err := r.db.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("задача не найдена")
		}
		return nil, err
	}
	return &task, nil
}

// CreateTask создает новую задачу
func (r *TaskRepository) CreateTask(task *models.Task) error {
	return r.db.Create(task).Error
}

// UpdateTask обновляет существующую задачу
func (r *TaskRepository) UpdateTask(task *models.Task) error {
	result := r.db.Save(task)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("задача не найдена")
	}
	return nil
}

// DeleteTask удаляет задачу по идентификатору
func (r *TaskRepository) DeleteTask(id uint) error {
	result := r.db.Delete(&models.Task{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("задача не найдена")
	}
	return nil
}


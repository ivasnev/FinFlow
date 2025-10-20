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
	var dbTasks []Task
	if err := r.db.Where("event_id = ?", eventID).Find(&dbTasks).Error; err != nil {
		return nil, err
	}
	return extractSlice(dbTasks), nil
}

// GetTaskByID возвращает задачу по идентификатору
func (r *TaskRepository) GetTaskByID(id uint) (*models.Task, error) {
	var dbTask Task
	if err := r.db.First(&dbTask, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("задача не найдена")
		}
		return nil, err
	}
	return extract(&dbTask), nil
}

// CreateTask создает новую задачу
func (r *TaskRepository) CreateTask(task *models.Task) error {
	dbTask := load(task)
	if err := r.db.Create(dbTask).Error; err != nil {
		return err
	}
	task.ID = dbTask.ID
	return nil
}

// UpdateTask обновляет существующую задачу
func (r *TaskRepository) UpdateTask(task *models.Task) error {
	dbTask := load(task)
	result := r.db.Save(dbTask)
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
	result := r.db.Delete(&Task{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("задача не найдена")
	}
	return nil
}

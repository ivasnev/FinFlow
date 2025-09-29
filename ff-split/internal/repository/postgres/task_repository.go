package postgres

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Task представляет модель задачи
type Task struct {
	ID          uint      `gorm:"primaryKey"`
	UserID      int64     `gorm:"column:user_id"`
	EventID     int64     `gorm:"column:event_id"`
	Title       string    `gorm:"column:title;not null"`
	Description string    `gorm:"column:description"`
	Priority    int       `gorm:"column:priority;default:0"`
	CreatedAt   time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
}

// TaskRepository интерфейс репозитория задач
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository создает новый репозиторий задач
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// GetTasksByEventID возвращает список всех задач мероприятия
func (r *TaskRepository) GetTasksByEventID(eventID int64) ([]Task, error) {
	var tasks []Task
	if err := r.db.Where("event_id = ?", eventID).Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTaskByID возвращает задачу по идентификатору
func (r *TaskRepository) GetTaskByID(id uint) (*Task, error) {
	var task Task
	if err := r.db.First(&task, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("задача не найдена")
		}
		return nil, err
	}
	return &task, nil
}

// CreateTask создает новую задачу
func (r *TaskRepository) CreateTask(task *Task) error {
	return r.db.Create(task).Error
}

// UpdateTask обновляет существующую задачу
func (r *TaskRepository) UpdateTask(task *Task) error {
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
	result := r.db.Delete(&Task{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("задача не найдена")
	}
	return nil
}

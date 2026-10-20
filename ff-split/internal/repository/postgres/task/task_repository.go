package task

import (
	"context"
	"errors"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository"
	"github.com/ivasnev/FinFlow/ff-split/internal/repository/postgres/history"
	"strconv"

	customErrors "github.com/ivasnev/FinFlow/ff-split/internal/common/errors"
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
			return nil, customErrors.NewEntityNotFoundError(strconv.Itoa(int(id)), "task")
		}
		return nil, err
	}
	return extract(&dbTask), nil
}

// WithContext привязывает запросы к автору и сроку жизни операции.
func (r *TaskRepository) WithContext(ctx context.Context) repository.Task {
	return NewTaskRepository(r.db.WithContext(ctx))
}

// CreateTask атомарно сохраняет задачу и запись журнала.
func (r *TaskRepository) CreateTask(task *models.Task) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if task.EventID != nil {
			if err := history.LockEvent(tx, *task.EventID); err != nil {
				return err
			}
		}
		row := load(task)
		if err := tx.Create(row).Error; err != nil {
			return err
		}
		task.ID = row.ID
		after, err := history.Task(tx, uint(row.ID))
		if err != nil {
			return err
		}
		if task.EventID != nil {
			return history.Record(tx.Statement.Context, tx, *task.EventID, "task.created", int64(task.ID), nil, after)
		}
		return nil
	})
}

// change удерживает событие и задачу до окончания изменения и записи журнала.
func (r *TaskRepository) change(id uint, task *models.Task) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var eventID int64
		if err := tx.Raw("SELECT event_id FROM tasks WHERE id=?", id).Scan(&eventID).Error; err != nil {
			return err
		}
		if err := history.LockEvent(tx, eventID); err != nil {
			return err
		}
		var current Task
		if err := tx.First(&current, id).Error; err != nil {
			return err
		}
		before, err := history.Task(tx, id)
		if err != nil {
			return err
		}
		kind := "task.deleted"
		var after map[string]string
		if task == nil {
			if err := tx.Delete(&Task{}, id).Error; err != nil {
				return err
			}
		} else {
			row := load(task)
			if err := tx.Model(&Task{}).Where("id=?", id).Select("title", "description", "user_id", "priority").Updates(row).Error; err != nil {
				return err
			}
			after, err = history.Task(tx, id)
			if err != nil {
				return err
			}
			kind = "task.updated"
		}
		return history.Record(tx.Statement.Context, tx, eventID, kind, int64(id), before, after)
	})
}

// UpdateTask сохраняет новую задачу вместе со снимком прежних полей.
func (r *TaskRepository) UpdateTask(task *models.Task) error { return r.change(uint(task.ID), task) }

// DeleteTask сохраняет снимок перед удалением задачи.
func (r *TaskRepository) DeleteTask(id uint) error { return r.change(id, nil) }

package task

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// extract преобразует модель задачи БД в бизнес-модель
func extract(dbTask *Task) *models.Task {
	if dbTask == nil {
		return nil
	}

	return &models.Task{
		ID:          dbTask.ID,
		UserID:      dbTask.UserID,
		EventID:     dbTask.EventID,
		Title:       dbTask.Title,
		Description: dbTask.Description,
		Priority:    dbTask.Priority,
		CreatedAt:   dbTask.CreatedAt,
	}
}

// extractSlice преобразует слайс моделей задач БД в бизнес-модели
func extractSlice(dbTasks []Task) []models.Task {
	tasks := make([]models.Task, len(dbTasks))
	for i, dbTask := range dbTasks {
		if extracted := extract(&dbTask); extracted != nil {
			tasks[i] = *extracted
		}
	}
	return tasks
}

// load преобразует бизнес-модель задачи в модель БД
func load(task *models.Task) *Task {
	if task == nil {
		return nil
	}

	return &Task{
		ID:          task.ID,
		UserID:      task.UserID,
		EventID:     task.EventID,
		Title:       task.Title,
		Description: task.Description,
		Priority:    task.Priority,
		CreatedAt:   task.CreatedAt,
	}
}

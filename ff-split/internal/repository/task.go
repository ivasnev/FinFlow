package repository

import (
	"github.com/ivasnev/FinFlow/ff-split/internal/models"
)

// Task определяет методы для работы с задачами
type Task interface {
	GetTasksByEventID(eventID int64) ([]models.Task, error)
	GetTaskByID(id uint) (*models.Task, error)
	CreateTask(task *models.Task) error
	UpdateTask(task *models.Task) error
	DeleteTask(id uint) error
}

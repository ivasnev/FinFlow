package service

import (
	"context"

	"github.com/ivasnev/FinFlow/ff-split/internal/api/dto"
)

// Task определяет методы для работы с задачами
type Task interface {
	GetTasksByEventID(ctx context.Context, eventID int64) ([]dto.TaskDTO, error)
	GetTaskByID(ctx context.Context, id uint) (*dto.TaskDTO, error)
	CreateTask(ctx context.Context, eventID int64, taskRequest *dto.TaskRequest) (*dto.TaskDTO, error)
	UpdateTask(ctx context.Context, id uint, taskRequest *dto.TaskRequest) (*dto.TaskDTO, error)
	DeleteTask(ctx context.Context, id uint) error
}


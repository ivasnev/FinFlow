package dto

import "time"

// TaskDTO представляет DTO для задачи
type TaskDTO struct {
	ID          uint      `json:"id"`
	UserID      int64     `json:"user_id"`
	EventID     int64     `json:"event_id"`
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
}

// TaskRequest представляет запрос на создание/обновление задачи
type TaskRequest struct {
	UserID      int64  `json:"user_id" binding:"required"`
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

// TaskResponse представляет ответ на операцию с задачей
type TaskResponse struct {
	Task TaskDTO `json:"task"`
}

// TaskListResponse представляет ответ со списком задач
type TaskListResponse struct {
	Tasks []TaskDTO `json:"tasks"`
}

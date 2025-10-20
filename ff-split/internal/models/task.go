package models

import "time"

// Task представляет задачу в системе
type Task struct {
	ID          int
	UserID      *int64
	EventID     *int64
	Title       string
	Description string
	Priority    int
	CreatedAt   time.Time

	// Отношения
	User  *User
	Event *Event
}

package models

import "time"

// Activity представляет действие в системе
type Activity struct {
	ID          int
	EventID     *int64
	UserID      *int64
	Description string
	IconID      int
	CreatedAt   time.Time

	// Отношения
	Icon  *Icon
	Event *Event
	User  *User
}

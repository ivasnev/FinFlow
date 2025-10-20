package models

// Event представляет модель мероприятия
type Event struct {
	ID          int64
	Name        string
	Description string
	CategoryID  *int
	ImageID     string
	Status      string

	// Отношения
	Category     *EventCategory
	Users        []User
	Activities   []Activity
	Transactions []Transaction
	Tasks        []Task
}

// EventCategory представляет категорию мероприятия
type EventCategory struct {
	ID     int
	Name   string
	IconID int

	// Отношения
	Icon   *Icon
	Events []Event
}

// UserEvent представляет связь между пользователем и мероприятием
type UserEvent struct {
	UserID  int64
	EventID int64
}

package event

// Event представляет модель мероприятия в БД
type Event struct {
	ID          int64  `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string `gorm:"column:name;not null"`
	Description string `gorm:"column:description"`
	CategoryID  *int   `gorm:"column:category_id"`
	ImageID     string `gorm:"column:image_id"`
	Status      string `gorm:"column:status;default:active"`
}

// TableName задает имя таблицы для модели Event
func (Event) TableName() string {
	return "events"
}

// EventCategory представляет категорию мероприятия в БД
type EventCategory struct {
	ID     int    `gorm:"column:id;primaryKey;autoIncrement"`
	Name   string `gorm:"column:name;not null"`
	IconID int    `gorm:"column:icon_id"`
}

// TableName задает имя таблицы для модели EventCategory
func (EventCategory) TableName() string {
	return "event_categories"
}

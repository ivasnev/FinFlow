package user

// User представляет модель пользователя в БД
type User struct {
	ID              int64  `gorm:"column:id;primaryKey;autoIncrement"`
	UserID          *int64 `gorm:"column:user_id;uniqueIndex"`
	NicknameCashed  string `gorm:"column:nickname_cashed"`
	NameCashed      string `gorm:"column:name_cashed"`
	PhotoUUIDCashed string `gorm:"column:photo_uuid_cashed"`
	IsDummy         bool   `gorm:"column:is_dummy;default:false"`
}

// TableName задает имя таблицы для модели User
func (User) TableName() string {
	return "users"
}

// UserEvent представляет связь между пользователем и мероприятием в БД
type UserEvent struct {
	UserID  int64 `gorm:"column:user_id;primaryKey"`
	EventID int64 `gorm:"column:event_id;primaryKey"`
}

// TableName задает имя таблицы для модели UserEvent
func (UserEvent) TableName() string {
	return "user_event"
}

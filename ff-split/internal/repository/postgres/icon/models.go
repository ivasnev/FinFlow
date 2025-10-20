package icon

// Icon представляет иконку для типа транзакции в БД
type Icon struct {
	ID       int    `gorm:"column:id;primaryKey"`
	Name     string `gorm:"column:name;not null"`
	FileUUID string `gorm:"column:file_uuid;not null"`
}

// TableName задает имя таблицы для модели Icon
func (Icon) TableName() string {
	return "icons"
}

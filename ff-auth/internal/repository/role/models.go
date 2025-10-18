package role

// RoleEntity представляет роль в системе
type RoleEntity struct {
	ID   int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name string `gorm:"type:text;unique;not null;column:name" json:"name"`
}

// TableName устанавливает имя таблицы для модели RoleEntity
func (RoleEntity) TableName() string {
	return "roles"
}

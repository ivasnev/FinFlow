package models

import "time"

// HistoryEntry сохраняет имена и значения на момент изменения, включая удалённые расходы.
type HistoryEntry struct {
	ID           int64             `json:"id"`
	EventID      int64             `json:"event_id"`
	ActorID      int64             `json:"actor_id"`
	ActorName    string            `json:"actor_name"`
	Kind         string            `json:"kind"`
	EntityID     int64             `json:"entity_id"`
	BeforeValues map[string]string `json:"before" gorm:"serializer:json"`
	AfterValues  map[string]string `json:"after" gorm:"serializer:json"`
	CreatedAt    time.Time         `json:"created_at"`
}

// TableName фиксирует имя неизменяемого журнала.
func (HistoryEntry) TableName() string { return "event_history" }

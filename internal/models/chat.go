package models

import (
	"time"
)

// ChatDB представляет модель чата для хранения в базе данных.
type ChatDB struct {
	ChatUUID      string    `json:"chat_uuid" db:"chat_uuid"`             // уникальный идентификатор чата
	CreatedByUUID string    `json:"created_by_uuid" db:"created_by_uuid"` // пользователь, который создал чат
	CreatedAt     time.Time `json:"created_at" db:"created_at"`           // дата создания записи
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`           // дата последнего обновления записи
}

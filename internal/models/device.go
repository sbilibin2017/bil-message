package models

import (
	"time"

	"github.com/google/uuid"
)

// DeviceDB представляет модель устройства для хранения в базе данных.
type DeviceDB struct {
	DeviceUUID uuid.UUID `json:"device_uuid" db:"device_uuid"` // уникальный идентификатор устройства
	UserUUID   uuid.UUID `json:"user_uuid" db:"user_uuid"`     // идентификатор пользователя-владельца
	PublicKey  string    `json:"public_key" db:"public_key"`   // публичный ключ клиента
	CreatedAt  time.Time `json:"created_at" db:"created_at"`   // дата создания записи
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`   // дата последнего обновления записи
}

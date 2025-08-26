package models

import (
	"time"

	"github.com/google/uuid"
)

// DeviceDB представляет устройство в бд.
type DeviceDB struct {
	DeviceUUID uuid.UUID `json:"device_uuid" db:"device_uuid"`
	UserUUID   uuid.UUID `json:"user_uuid" db:"user_uuid"`
	PublicKey  string    `json:"public_key" db:"public_key"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

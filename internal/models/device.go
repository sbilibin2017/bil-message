package models

import (
	"time"

	"github.com/google/uuid"
)

type DeviceDB struct {
	DeviceUUID uuid.UUID `db:"device_uuid"`
	UserUUID   uuid.UUID `db:"user_uuid"`
	PublicKey  string    `db:"public_key"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

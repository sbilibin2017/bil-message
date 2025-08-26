package models

import (
	"time"

	"github.com/google/uuid"
)

// RoomDB представляет комнату в БД.
type RoomDB struct {
	RoomUUID  uuid.UUID `json:"room_uuid" db:"room_uuid"`
	OwnerUUID uuid.UUID `json:"owner_uuid" db:"owner_uuid"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type RoomMessage struct {
	RoomUUID  uuid.UUID `json:"room_uuid"`
	UserUUID  uuid.UUID `json:"user_uuid"`
	Message   string    `json:"message"`
	Timestamp int64     `json:"timestamp"`
}

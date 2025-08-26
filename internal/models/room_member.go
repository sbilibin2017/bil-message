package models

import (
	"time"

	"github.com/google/uuid"
)

// RoomMemberDB представляет участника комнаты в бд.
type RoomMemberDB struct {
	RoomUUID   uuid.UUID `json:"room_uuid" db:"room_uuid"`
	MemberUUID uuid.UUID `json:"member_uuid" db:"member_uuid"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

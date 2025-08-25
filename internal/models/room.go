package models

import (
	"time"

	"github.com/google/uuid"
)

// RoomDB представляет запись в таблице rooms
type RoomDB struct {
	RoomUUID    uuid.UUID `json:"room_uuid" db:"room_uuid"`       // UUID комнаты (PK)
	CreatorUUID uuid.UUID `json:"creator_uuid" db:"creator_uuid"` // UUID создателя комнаты (FK)
	CreatedAt   time.Time `json:"created_at" db:"created_at"`     // Время создания комнаты
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`     // Время последнего обновления комнаты
}

// RoomMemberDB представляет запись участника комнаты в таблице room_members
type RoomMemberDB struct {
	RoomUUID  uuid.UUID `json:"room_uuid" db:"room_uuid"`   // UUID комнаты (FK)
	UserUUID  uuid.UUID `json:"user_uuid" db:"user_uuid"`   // UUID пользователя (FK)
	JoinedAt  time.Time `json:"joined_at" db:"joined_at"`   // Время присоединения к комнате
	CreatedAt time.Time `json:"created_at" db:"created_at"` // Время создания записи
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Время последнего обновления записи
}

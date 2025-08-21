package models

import (
	"time"
)

// ChatMemberDB представляет модель участника чата для хранения в базе данных.
type ChatMemberDB struct {
	ChatMemberUUID string    `json:"chat_member_uuid" db:"chat_member_uuid"` // уникальный идентификатор записи участника
	ChatUUID       string    `json:"chat_uuid" db:"chat_uuid"`               // идентификатор чата
	UserUUID       string    `json:"user_uuid" db:"user_uuid"`               // идентификатор пользователя
	JoinedAt       time.Time `json:"joined_at" db:"joined_at"`               // дата присоединения к чату
	CreatedAt      time.Time `json:"created_at" db:"created_at"`             // дата создания записи
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`             // дата последнего обновления записи
}

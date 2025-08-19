package models

import (
	"time"

	"github.com/google/uuid"
)

// UserDB представляет модель пользователя для хранения в базе данных.
type UserDB struct {
	UserUUID     uuid.UUID `json:"user_uuid" db:"user_uuid"`         // уникальный идентификатор пользователя
	Username     string    `json:"username" db:"username"`           // имя пользователя
	PasswordHash string    `json:"password_hash" db:"password_hash"` // хэш пароля, не отдаётся в JSON
	CreatedAt    time.Time `json:"created_at" db:"created_at"`       // дата создания записи
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`       // дата последнего обновления записи
}

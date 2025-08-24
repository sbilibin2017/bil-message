package models

import (
	"time"

	"github.com/google/uuid"
)

// RegisterRequest — тело запроса для регистрации нового пользователя
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterResponse — ответ сервера при регистрации
type RegisterResponse struct {
	UserUUID string `json:"user_uuid"` // UUID нового пользователя
}

// UserDB представляет запись в таблице users
type UserDB struct {
	UserUUID     uuid.UUID `json:"user_uuid" db:"user_uuid"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"password_hash" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserDeviceDB представляет запись устройства пользователя в базе
type UserDeviceDB struct {
	DeviceUUID uuid.UUID `json:"device_uuid" db:"device_uuid"` // UUID устройства (PK)
	UserUUID   uuid.UUID `json:"user_uuid" db:"user_uuid"`     // UUID пользователя (FK)
	PublicKey  string    `json:"public_key" db:"public_key"`   // Публичный ключ устройства
	CreatedAt  time.Time `json:"created_at" db:"created_at"`   // Время создания записи
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`   // Время последнего обновления записи
}

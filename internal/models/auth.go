package models

import "github.com/google/uuid"

// TokenPayload хранит данные, которые мы закодировали в JWT
type TokenPayload struct {
	UserUUID   uuid.UUID `json:"user_uuid"`   // UUID пользователя
	ClientUUID uuid.UUID `json:"client_uuid"` // UUID устройства/клиента
}

package models

import "github.com/golang-jwt/jwt/v5"

// TokenPayload представляет полезную нагрузку (payload) токена.
type TokenPayload struct {
	UserUUID   string `json:"user_uuid"`   // идентификатор пользователя в системе.
	DeviceUUID string `json:"device_uuid"` // идентификатор устройства.
}

// Claims — структура для хранения полезной нагрузки токена (включая служебные поля JWT).
type Claims struct {
	UserUUID   string `json:"user_uuid"`   // идентификатор пользователя в системе.
	DeviceUUID string `json:"device_uuid"` // идентификатор устройства.
	jwt.RegisteredClaims
}

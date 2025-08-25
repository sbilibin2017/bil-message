package models

type TokenResponse struct {
	UserUUID   string `json:"user_uuid"`
	DeviceUUID string `json:"device_uuid"`
}

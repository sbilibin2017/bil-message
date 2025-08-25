package models

type Message struct {
	RoomUUID   string `json:"room_uuid"`
	UserUUID   string `json:"user_uuid"`
	DeviceUUID string `json:"device_uuid"`
	Text       string `json:"text"`
	Timestamp  int64  `json:"timestamp"`
}

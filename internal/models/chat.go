package models

import "time"

// ChatDB представляет модель чата для хранения в базе данных.
type ChatDB struct {
	ChatUUID          string    `db:"chat_uuid"`
	ParticipantsUUIDs string    `db:"participants_uuids"`
	CreatedByUUID     string    `db:"created_by_uuid"`
	CreatedAt         time.Time `db:"created_at"`
	UpdatedAt         time.Time `db:"updated_at"`
}

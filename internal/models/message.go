package models

import "time"

// MessageDB представляет модель сообщения для хранения в базе данных.
type MessageDB struct {
	MessageUUID   string    `json:"message_uuid" db:"message_uuid"`     // уникальный идентификатор сообщения
	ChatUUID      string    `json:"chat_uuid" db:"chat_uuid"`           // идентификатор чата
	SenderUUID    string    `json:"sender_uuid" db:"sender_uuid"`       // идентификатор пользователя-отправителя
	EncryptedText string    `json:"encrypted_text" db:"encrypted_text"` // зашифрованный текст сообщения
	CreatedAt     time.Time `json:"created_at" db:"created_at"`         // дата создания записи
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`         // дата последнего обновления записи
}

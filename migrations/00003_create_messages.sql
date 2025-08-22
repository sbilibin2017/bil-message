-- +goose Up
CREATE TABLE messages (
    message_uuid TEXT PRIMARY KEY,
    chat_uuid TEXT NOT NULL,
    sender_uuid TEXT NOT NULL,
    encrypted_text TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chat_uuid) REFERENCES chats(chat_uuid),
    FOREIGN KEY (sender_uuid) REFERENCES users(user_uuid)
);

-- +goose Down
DROP TABLE IF EXISTS messages;

-- +goose Up
CREATE TABLE chats (
    chat_uuid TEXT PRIMARY KEY,
    created_by_uuid TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS chats;

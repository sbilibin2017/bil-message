-- +goose Up
CREATE TABLE chats (
    chat_uuid TEXT PRIMARY KEY,    
    participants_uuids TEXT NOT NULL, 
    created_by_uuid TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (created_by_uuid) REFERENCES users(user_uuid)
);

-- +goose Down
DROP TABLE IF EXISTS chats;
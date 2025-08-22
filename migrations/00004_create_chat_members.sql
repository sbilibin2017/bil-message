-- +goose Up
CREATE TABLE chat_members (
    chat_member_uuid TEXT PRIMARY KEY,
    chat_uuid TEXT NOT NULL,
    user_uuid TEXT NOT NULL,    
    joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(chat_uuid, user_uuid)
);

-- +goose Down
DROP TABLE IF EXISTS chat_members;

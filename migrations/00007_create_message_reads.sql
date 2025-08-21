-- +goose Up
CREATE TABLE message_reads (
    message_read_uuid TEXT PRIMARY KEY,
    message_uuid TEXT NOT NULL,
    user_uuid TEXT NOT NULL,
    read_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS message_reads;

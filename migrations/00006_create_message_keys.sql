-- +goose Up
CREATE TABLE message_keys (
    message_key_uuid TEXT PRIMARY KEY,
    message_uuid TEXT NOT NULL,
    device_uuid TEXT NOT NULL,
    encrypted_symmetric_key TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS message_keys;

-- +goose Up
CREATE TABLE message_keys (
    message_key_uuid UUID PRIMARY KEY,
    message_id UUID NOT NULL REFERENCES messages(message_uuid) ON DELETE CASCADE,
    recipient_id UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    encrypted_symmetric_key TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS message_keys;

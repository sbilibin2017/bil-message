-- +goose Up
CREATE TABLE message_reads (
    message_read_uuid UUID PRIMARY KEY,
    message_uuid UUID NOT NULL REFERENCES messages(message_uuid) ON DELETE CASCADE,
    user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    read_at TIMESTAMP DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS message_reads;
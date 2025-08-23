-- +goose Up
CREATE TABLE room_messages (
    message_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_uuid    UUID NOT NULL REFERENCES rooms(room_uuid) ON DELETE CASCADE,
    sender_uuid  UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    ciphertext   TEXT NOT NULL,
    sent_at      TIMESTAMP DEFAULT NOW(),
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS room_messages;

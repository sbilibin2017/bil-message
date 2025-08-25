-- +migrate Up
CREATE TABLE messages (
    message_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_uuid UUID NOT NULL REFERENCES rooms(room_uuid) ON DELETE CASCADE,
    sender_user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    encrypted_content BYTEA NOT NULL,   -- зашифрованное сообщение
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS messages;
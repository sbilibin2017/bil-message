-- +goose Up
CREATE TABLE messages (
    message_uuid UUID PRIMARY KEY,
    chat_id UUID NOT NULL REFERENCES chats(chat_uuid) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    encrypted_text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now() NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS messages;

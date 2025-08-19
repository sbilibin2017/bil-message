
-- +goose Up
CREATE TABLE chat_members (
    chat_member_uuid UUID PRIMARY KEY,
    chat_uuid UUID NOT NULL REFERENCES chats(chat_uuid) ON DELETE CASCADE,
    user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    role VARCHAR(20),
    joined_at TIMESTAMP DEFAULT now() NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS chat_members;
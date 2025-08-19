-- +goose Up
CREATE TABLE chats (
    chat_uuid UUID PRIMARY KEY,
    name VARCHAR(100),
    type_uuid UUID NOT NULL REFERENCES chat_types(chat_type_uuid),
    created_by_uuid UUID NOT NULL REFERENCES users(user_uuid),
    created_at TIMESTAMP DEFAULT now() NOT NULL,
    updated_at TIMESTAMP DEFAULT now() NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS chats;
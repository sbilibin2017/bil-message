
-- +goose Up
CREATE TABLE chat_types (
    chat_type_uuid UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT
);

-- +goose Down
DROP TABLE IF EXISTS chat_types;
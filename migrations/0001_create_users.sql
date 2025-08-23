-- +goose Up
CREATE TABLE users (
    user_uuid     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS users;

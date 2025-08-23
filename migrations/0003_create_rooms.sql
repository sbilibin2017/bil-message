-- +goose Up
CREATE TABLE rooms (
    room_uuid    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(255) NOT NULL,
    creator_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS rooms;

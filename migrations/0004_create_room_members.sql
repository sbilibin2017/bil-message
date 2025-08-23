-- +goose Up
CREATE TABLE room_members (
    room_uuid  UUID NOT NULL REFERENCES rooms(room_uuid) ON DELETE CASCADE,
    user_uuid  UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    role       VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at  TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (room_uuid, user_uuid)
);

-- +goose Down
DROP TABLE IF EXISTS room_members;

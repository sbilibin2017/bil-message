-- +goose Up
CREATE TABLE room_members (
    room_uuid UUID NOT NULL REFERENCES rooms(room_uuid) ON DELETE CASCADE,
    member_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    PRIMARY KEY (room_uuid, member_uuid)
);

-- +goose Down
DROP TABLE IF EXISTS room_members;

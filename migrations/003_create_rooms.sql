-- +migrate Up
CREATE TABLE rooms (
    room_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS rooms;
-- +goose Up
CREATE TABLE rooms (
    room_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),    
    owner_uuid UUID NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS rooms;

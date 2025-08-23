-- +goose Up
CREATE TABLE room_keys (
    room_uuid     UUID NOT NULL REFERENCES rooms(room_uuid) ON DELETE CASCADE,
    device_uuid   UUID NOT NULL REFERENCES user_devices(device_uuid) ON DELETE CASCADE,
    encrypted_key TEXT NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (room_uuid, device_uuid)
);

-- +goose Down
DROP TABLE IF EXISTS room_keys;

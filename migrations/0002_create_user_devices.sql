-- +goose Up
CREATE TABLE user_devices (
    device_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_uuid   UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    public_key  TEXT NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS user_devices;

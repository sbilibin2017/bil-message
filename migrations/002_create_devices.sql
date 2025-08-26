-- +migrate Up
CREATE TABLE devices (
    device_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_uuid UUID NOT NULL REFERENCES users(user_uuid) ON DELETE CASCADE,
    public_key TEXT NOT NULL, -- публичный ключ устройства для E2EE
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS devices;
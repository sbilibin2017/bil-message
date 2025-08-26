-- +goose Up
CREATE TABLE message_keys (
    message_key_uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_uuid UUID NOT NULL REFERENCES messages(message_uuid) ON DELETE CASCADE,
    device_uuid UUID NOT NULL REFERENCES devices(device_uuid) ON DELETE CASCADE,
    encrypted_symmetric_key BYTEA NOT NULL, -- симметричный ключ зашифрованный публичным ключом устройства
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS message_keys;
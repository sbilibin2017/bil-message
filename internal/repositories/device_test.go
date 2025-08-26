package repositories_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/repositories"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func setupDeviceTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	assert.NoError(t, err)

	schema := `
	CREATE TABLE devices (
		device_uuid TEXT PRIMARY KEY,
		user_uuid TEXT NOT NULL,
		public_key TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	assert.NoError(t, err)

	return db
}

func TestDeviceWriteAndReadRepository(t *testing.T) {
	ctx := context.Background()
	db := setupDeviceTestDB(t)
	defer db.Close()

	deviceWriteRepo := repositories.NewDeviceWriteRepository(db)
	deviceReadRepo := repositories.NewDeviceReadRepository(db)

	deviceUUID := uuid.New()
	userUUID := uuid.New()
	publicKey := "pubkey123"

	// --- Test Save ---
	err := deviceWriteRepo.Save(ctx, deviceUUID, userUUID, publicKey)
	assert.NoError(t, err)

	// --- Test Get ---
	device, err := deviceReadRepo.Get(ctx, deviceUUID)
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, deviceUUID, device.DeviceUUID)
	assert.Equal(t, userUUID, device.UserUUID)
	assert.Equal(t, publicKey, device.PublicKey)

	// --- Test Update ---
	newKey := "newpubkey456"
	err = deviceWriteRepo.Save(ctx, deviceUUID, userUUID, newKey)
	assert.NoError(t, err)

	deviceUpdated, err := deviceReadRepo.Get(ctx, deviceUUID)
	assert.NoError(t, err)
	assert.Equal(t, newKey, deviceUpdated.PublicKey)

	// --- Test Get non-existent device ---
	nonDevice, err := deviceReadRepo.Get(ctx, uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, nonDevice)
}

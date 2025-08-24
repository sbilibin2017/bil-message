package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/repositories"
	"github.com/stretchr/testify/assert"

	_ "modernc.org/sqlite"
)

func setupDeviceDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	assert.NoError(t, err)

	schema := `
	CREATE TABLE user_devices (
		device_uuid TEXT PRIMARY KEY,
		user_uuid   TEXT NOT NULL,
		public_key  TEXT NOT NULL,
		created_at  DATETIME NOT NULL,
		updated_at  DATETIME NOT NULL
	);`
	_, err = db.Exec(schema)
	assert.NoError(t, err)

	return db
}

func TestDeviceWriteAndRead(t *testing.T) {
	db := setupDeviceDB(t)
	writeRepo := repositories.NewDeviceWriteRepository(db)
	readRepo := repositories.NewDeviceReadRepository(db)
	ctx := context.Background()

	deviceUUID := uuid.New()
	userUUID := uuid.New()
	publicKey := "pubkey123"

	// Save new device
	err := writeRepo.Save(ctx, deviceUUID, userUUID, publicKey)
	assert.NoError(t, err)

	// Get by deviceUUID
	device, err := readRepo.Get(ctx, deviceUUID)
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, deviceUUID, device.DeviceUUID)
	assert.Equal(t, userUUID, device.UserUUID)
	assert.Equal(t, publicKey, device.PublicKey)

	// Save with same deviceUUID but updated publicKey
	newPublicKey := "newpubkey456"
	time.Sleep(time.Millisecond * 10) // ensure updated_at will be different
	err = writeRepo.Save(ctx, deviceUUID, userUUID, newPublicKey)
	assert.NoError(t, err)

	// Read again
	device, err = readRepo.Get(ctx, deviceUUID)
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, deviceUUID, device.DeviceUUID)
	assert.Equal(t, userUUID, device.UserUUID)
	assert.Equal(t, newPublicKey, device.PublicKey)
}

func TestGetNonExistingDevice(t *testing.T) {
	db := setupDeviceDB(t)
	readRepo := repositories.NewDeviceReadRepository(db)
	ctx := context.Background()

	device, err := readRepo.Get(ctx, uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, device)
}

func TestSaveMultipleDevices(t *testing.T) {
	db := setupDeviceDB(t)
	writeRepo := repositories.NewDeviceWriteRepository(db)
	readRepo := repositories.NewDeviceReadRepository(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		deviceUUID := uuid.New()
		userUUID := uuid.New()
		publicKey := "pub" + string(rune(i+'0'))

		err := writeRepo.Save(ctx, deviceUUID, userUUID, publicKey)
		assert.NoError(t, err)

		device, err := readRepo.Get(ctx, deviceUUID)
		assert.NoError(t, err)
		assert.NotNil(t, device)
		assert.Equal(t, publicKey, device.PublicKey)
		assert.Equal(t, deviceUUID, device.DeviceUUID)
		assert.Equal(t, userUUID, device.UserUUID)
	}
}

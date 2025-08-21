package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/sbilibin2017/bil-message/internal/models"
	_ "modernc.org/sqlite"
)

func setupDeviceTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect to sqlite: %v", err)
	}

	schema := `
	CREATE TABLE devices (
		device_uuid TEXT PRIMARY KEY,
		user_uuid TEXT NOT NULL,
		public_key TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create devices table: %v", err)
	}

	return db
}

func TestDeviceWriteAndRead(t *testing.T) {
	ctx := context.Background()
	db := setupDeviceTestDB(t)
	defer db.Close()

	writeRepo := NewDeviceWriteRepository(db)
	readRepo := NewDeviceReadRepository(db)

	deviceUUID := uuid.New().String()
	userUUID := uuid.New().String()
	publicKey := "pubkey123"

	device := &models.DeviceDB{
		DeviceUUID: deviceUUID,
		UserUUID:   userUUID,
		PublicKey:  publicKey,
	}

	// Save device
	err := writeRepo.Save(ctx, device)
	assert.NoError(t, err)

	// Get device
	got, err := readRepo.GetByUUID(ctx, deviceUUID)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, deviceUUID, got.DeviceUUID)
	assert.Equal(t, userUUID, got.UserUUID)
	assert.Equal(t, publicKey, got.PublicKey)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
}

func TestDeviceWrite_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	db := setupDeviceTestDB(t)
	defer db.Close()

	writeRepo := NewDeviceWriteRepository(db)
	readRepo := NewDeviceReadRepository(db)

	deviceUUID := uuid.New().String()
	userUUID := uuid.New().String()
	publicKey1 := "pubkey1"
	publicKey2 := "pubkey2"

	device := &models.DeviceDB{
		DeviceUUID: deviceUUID,
		UserUUID:   userUUID,
		PublicKey:  publicKey1,
	}

	// First save
	err := writeRepo.Save(ctx, device)
	assert.NoError(t, err)

	// Update existing device
	device.PublicKey = publicKey2
	err = writeRepo.Save(ctx, device)
	assert.NoError(t, err)

	got, err := readRepo.GetByUUID(ctx, deviceUUID)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, publicKey2, got.PublicKey)
}

func TestDeviceRead_NotFound(t *testing.T) {
	ctx := context.Background()
	db := setupDeviceTestDB(t)
	defer db.Close()

	readRepo := NewDeviceReadRepository(db)

	got, err := readRepo.GetByUUID(ctx, uuid.New().String())
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestDeviceRead_GetError(t *testing.T) {
	ctx := context.Background()
	db := setupDeviceTestDB(t)
	defer db.Close()

	readRepo := NewDeviceReadRepository(db)

	// Ломаем таблицу
	_, err := db.Exec(`DROP TABLE devices`)
	assert.NoError(t, err)

	got, err := readRepo.GetByUUID(ctx, uuid.New().String())
	assert.Error(t, err)
	assert.Nil(t, got)
}

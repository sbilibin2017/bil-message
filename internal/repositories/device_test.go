package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	_ "modernc.org/sqlite"
)

func setupClientTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect to sqlite: %v", err)
	}

	schema := `
	CREATE TABLE devices (
		device_uuid TEXT PRIMARY KEY,
		user_uuid TEXT NOT NULL,
		public_key TEXT NOT NULL,
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

func TestClientWriteAndRead(t *testing.T) {
	ctx := context.Background()
	db := setupClientTestDB(t)
	defer db.Close()

	writeRepo := NewDeviceWriteRepository(db)
	readRepo := NewDeviceReadRepository(db)

	clientUUID := uuid.New()
	userUUID := uuid.New()
	publicKey := "pubkey123"

	// Save client
	err := writeRepo.Save(ctx, clientUUID, userUUID, publicKey)
	assert.NoError(t, err)

	// Get client
	client, err := readRepo.GetByUUID(ctx, clientUUID)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, clientUUID, client.DeviceUUID)
	assert.Equal(t, userUUID, client.UserUUID)
	assert.Equal(t, publicKey, client.PublicKey)

	// Проверяем, что CreatedAt и UpdatedAt проставлены
	assert.WithinDuration(t, time.Now(), client.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), client.UpdatedAt, time.Second)
}

func TestClientWrite_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	db := setupClientTestDB(t)
	defer db.Close()

	writeRepo := NewDeviceWriteRepository(db)
	readRepo := NewDeviceReadRepository(db)

	clientUUID := uuid.New()
	userUUID := uuid.New()
	publicKey1 := "pubkey1"
	publicKey2 := "pubkey2"

	// First save
	err := writeRepo.Save(ctx, clientUUID, userUUID, publicKey1)
	assert.NoError(t, err)

	// Update existing client
	err = writeRepo.Save(ctx, clientUUID, userUUID, publicKey2)
	assert.NoError(t, err)

	client, err := readRepo.GetByUUID(ctx, clientUUID)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, publicKey2, client.PublicKey)
}

func TestClientRead_NotFound(t *testing.T) {
	ctx := context.Background()
	db := setupClientTestDB(t)
	defer db.Close()

	readRepo := NewDeviceReadRepository(db)

	client, err := readRepo.GetByUUID(ctx, uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, client) // теперь должно быть nil вместо ошибки
}

func TestClientRead_GetError(t *testing.T) {
	ctx := context.Background()
	db := setupClientTestDB(t)
	defer db.Close()

	readRepo := NewDeviceReadRepository(db)

	// Intentionally drop the table to simulate a SQL error
	_, err := db.Exec(`DROP TABLE devices`)
	assert.NoError(t, err)

	client, err := readRepo.GetByUUID(ctx, uuid.New())
	assert.Error(t, err)  // Should return a real SQL error
	assert.Nil(t, client) // Client should be nil on error
}

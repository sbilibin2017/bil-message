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

func setupRoomTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	assert.NoError(t, err)

	schema := `
	CREATE TABLE rooms (
		room_uuid TEXT PRIMARY KEY,
		owner_uuid TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	assert.NoError(t, err)

	return db
}

func TestRoomWriteAndReadRepository(t *testing.T) {
	ctx := context.Background()
	db := setupRoomTestDB(t)
	defer db.Close()

	roomWriteRepo := repositories.NewRoomWriteRepository(db)
	roomReadRepo := repositories.NewRoomReadRepository(db)

	roomUUID := uuid.New()
	ownerUUID := uuid.New()

	// --- Test Save ---
	err := roomWriteRepo.Save(ctx, roomUUID, ownerUUID)
	assert.NoError(t, err)

	// --- Test Get ---
	room, err := roomReadRepo.Get(ctx, roomUUID)
	assert.NoError(t, err)
	assert.NotNil(t, room)
	assert.Equal(t, roomUUID, room.RoomUUID)
	assert.Equal(t, ownerUUID, room.OwnerUUID)

	// --- Test Update (Save again should update updated_at) ---
	beforeUpdate := room.UpdatedAt
	time.Sleep(1 * time.Millisecond) // ensure updated_at changes
	newOwnerUUID := uuid.New()
	err = roomWriteRepo.Save(ctx, roomUUID, newOwnerUUID)
	assert.NoError(t, err)

	roomUpdated, err := roomReadRepo.Get(ctx, roomUUID)
	assert.NoError(t, err)
	assert.True(t, roomUpdated.UpdatedAt.After(beforeUpdate))
	assert.Equal(t, newOwnerUUID, roomUpdated.OwnerUUID)

	// --- Test Delete ---
	err = roomWriteRepo.Delete(ctx, roomUUID)
	assert.NoError(t, err)

	roomDeleted, err := roomReadRepo.Get(ctx, roomUUID)
	assert.NoError(t, err)
	assert.Nil(t, roomDeleted)
}

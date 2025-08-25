package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/repositories"
	"github.com/stretchr/testify/assert"

	_ "modernc.org/sqlite" // sqlite driver
)

func setupRoomDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	assert.NoError(t, err)

	schema := `
	CREATE TABLE rooms (
		room_uuid TEXT PRIMARY KEY,
		creator_uuid TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`
	_, err = db.Exec(schema)
	assert.NoError(t, err)

	return db
}

func TestRoomWriteAndRead(t *testing.T) {
	db := setupRoomDB(t)
	writeRepo := repositories.NewRoomWriteRepository(db)
	readRepo := repositories.NewRoomReadRepository(db)
	ctx := context.Background()

	roomUUID := uuid.New()
	creatorUUID := uuid.New()

	// Save new room
	err := writeRepo.Save(ctx, roomUUID, creatorUUID)
	assert.NoError(t, err)

	// Get by roomUUID
	room, err := readRepo.Get(ctx, roomUUID)
	assert.NoError(t, err)
	assert.NotNil(t, room)
	assert.Equal(t, roomUUID, room.RoomUUID)
	assert.Equal(t, creatorUUID, room.CreatorUUID)

	// Save with same room_uuid again (should update updated_at)
	oldUpdatedAt := room.UpdatedAt
	time.Sleep(1 * time.Millisecond) // ensure timestamp difference
	err = writeRepo.Save(ctx, roomUUID, creatorUUID)
	assert.NoError(t, err)

	room, err = readRepo.Get(ctx, roomUUID)
	assert.NoError(t, err)
	assert.NotNil(t, room)
	assert.Equal(t, roomUUID, room.RoomUUID)
	assert.Equal(t, creatorUUID, room.CreatorUUID)
	assert.True(t, room.UpdatedAt.After(oldUpdatedAt))
}

func TestGetNonExistingRoom(t *testing.T) {
	db := setupRoomDB(t)
	readRepo := repositories.NewRoomReadRepository(db)
	ctx := context.Background()

	room, err := readRepo.Get(ctx, uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, room)
}

func TestSaveMultipleRooms(t *testing.T) {
	db := setupRoomDB(t)
	writeRepo := repositories.NewRoomWriteRepository(db)
	readRepo := repositories.NewRoomReadRepository(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		roomUUID := uuid.New()
		creatorUUID := uuid.New()

		err := writeRepo.Save(ctx, roomUUID, creatorUUID)
		assert.NoError(t, err)

		room, err := readRepo.Get(ctx, roomUUID)
		assert.NoError(t, err)
		assert.NotNil(t, room)
		assert.Equal(t, roomUUID, room.RoomUUID)
		assert.Equal(t, creatorUUID, room.CreatorUUID)
	}
}

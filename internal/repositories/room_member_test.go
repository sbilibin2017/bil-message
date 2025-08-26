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

func setupRoomMembersTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	assert.NoError(t, err)

	schema := `
	CREATE TABLE room_members (
		room_uuid TEXT NOT NULL,
		member_uuid TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (room_uuid, member_uuid)
	);
	`
	_, err = db.Exec(schema)
	assert.NoError(t, err)

	return db
}

func TestRoomMemberWriteAndReadRepository(t *testing.T) {
	ctx := context.Background()
	db := setupRoomMembersTestDB(t)
	defer db.Close()

	writeRepo := repositories.NewRoomMemberWriteRepository(db)
	readRepo := repositories.NewRoomMemberReadRepository(db)

	roomUUID := uuid.New()
	memberUUID := uuid.New()

	// --- Test Save ---
	err := writeRepo.Save(ctx, roomUUID, memberUUID)
	assert.NoError(t, err)

	// --- Test Get ---
	member, err := readRepo.Get(ctx, roomUUID, memberUUID)
	assert.NoError(t, err)
	assert.NotNil(t, member)
	assert.Equal(t, roomUUID, member.RoomUUID)
	assert.Equal(t, memberUUID, member.MemberUUID)

	// --- Test Update (updated_at should change) ---
	time.Sleep(1 * time.Millisecond) // ensure updated_at changes
	err = writeRepo.Save(ctx, roomUUID, memberUUID)
	assert.NoError(t, err)

	memberUpdated, err := readRepo.Get(ctx, roomUUID, memberUUID)
	assert.NoError(t, err)
	assert.NotNil(t, memberUpdated)
	assert.Equal(t, roomUUID, memberUpdated.RoomUUID)
	assert.Equal(t, memberUUID, memberUpdated.MemberUUID)
	assert.True(t, memberUpdated.UpdatedAt.After(member.UpdatedAt))

	// --- Test Delete ---
	err = writeRepo.Delete(ctx, roomUUID, memberUUID)
	assert.NoError(t, err)

	memberDeleted, err := readRepo.Get(ctx, roomUUID, memberUUID)
	assert.NoError(t, err)
	assert.Nil(t, memberDeleted)

	// --- Test DeleteAllByRoom ---
	// Add multiple members
	member1 := uuid.New()
	member2 := uuid.New()
	_ = writeRepo.Save(ctx, roomUUID, member1)
	_ = writeRepo.Save(ctx, roomUUID, member2)

}

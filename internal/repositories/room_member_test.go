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

func setupRoomMembersDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	assert.NoError(t, err)

	schema := `
	CREATE TABLE room_members (
		room_uuid TEXT NOT NULL,
		user_uuid TEXT NOT NULL,
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		PRIMARY KEY (room_uuid, user_uuid)
	);`
	_, err = db.Exec(schema)
	assert.NoError(t, err)

	return db
}

func TestRoomMemberWriteAndRead(t *testing.T) {
	db := setupRoomMembersDB(t)
	writeRepo := repositories.NewRoomMemberWriteRepository(db)
	readRepo := repositories.NewRoomMemberReadRepository(db)
	ctx := context.Background()

	roomUUID := uuid.New()
	userUUID := uuid.New()

	// Save new member
	err := writeRepo.Save(ctx, roomUUID, userUUID, time.Time{})
	assert.NoError(t, err)

	// Get by roomUUID + userUUID
	member, err := readRepo.Get(ctx, roomUUID, userUUID)
	assert.NoError(t, err)
	assert.NotNil(t, member)
	assert.Equal(t, roomUUID, member.RoomUUID)
	assert.Equal(t, userUUID, member.UserUUID)

	// Save same member again (should update updated_at)
	oldUpdatedAt := member.UpdatedAt
	time.Sleep(1 * time.Millisecond) // чтобы timestamp точно изменился
	err = writeRepo.Save(ctx, roomUUID, userUUID, time.Time{})
	assert.NoError(t, err)

	member, err = readRepo.Get(ctx, roomUUID, userUUID)
	assert.NoError(t, err)
	assert.NotNil(t, member)
	assert.True(t, member.UpdatedAt.After(oldUpdatedAt))
}

func TestGetNonExistingMember(t *testing.T) {
	db := setupRoomMembersDB(t)
	readRepo := repositories.NewRoomMemberReadRepository(db)
	ctx := context.Background()

	member, err := readRepo.Get(ctx, uuid.New(), uuid.New())
	assert.NoError(t, err)
	assert.Nil(t, member)
}

func TestListMembers(t *testing.T) {
	db := setupRoomMembersDB(t)
	writeRepo := repositories.NewRoomMemberWriteRepository(db)
	readRepo := repositories.NewRoomMemberReadRepository(db)
	ctx := context.Background()

	roomUUID := uuid.New()
	userIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	for _, userUUID := range userIDs {
		err := writeRepo.Save(ctx, roomUUID, userUUID, time.Time{})
		assert.NoError(t, err)
	}

	members, err := readRepo.List(ctx, roomUUID)
	assert.NoError(t, err)
	assert.Len(t, members, len(userIDs))

	// проверяем, что все userUUID из userIDs есть в members
	memberMap := make(map[uuid.UUID]bool)
	for _, m := range members {
		memberMap[m.UserUUID] = true
	}

	for _, id := range userIDs {
		assert.True(t, memberMap[id], "userUUID %s not found in members", id)
	}
}

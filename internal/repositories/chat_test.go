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

func setupChatTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect to sqlite: %v", err)
	}

	schema := `
	CREATE TABLE chats (
		chat_uuid TEXT PRIMARY KEY,
		participants_uuids TEXT NOT NULL,
		created_by_uuid TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create chats table: %v", err)
	}

	return db
}

func TestChatWrite_Save(t *testing.T) {
	ctx := context.Background()
	db := setupChatTestDB(t)
	defer db.Close()

	repo := NewChatWriteRepository(db)

	chatUUID := uuid.New().String()
	createdByUUID := uuid.New().String()
	participants := createdByUUID + ",user2,user3"

	err := repo.Save(ctx, chatUUID, createdByUUID, participants)
	assert.NoError(t, err)

	var got models.ChatDB
	err = db.Get(&got, `
		SELECT chat_uuid, participants_uuids, created_by_uuid, created_at, updated_at
		FROM chats WHERE chat_uuid = ?`, chatUUID)
	assert.NoError(t, err)
	assert.Equal(t, chatUUID, got.ChatUUID)
	assert.Equal(t, createdByUUID, got.CreatedByUUID)
	assert.Equal(t, participants, got.ParticipantsUUIDs)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
}

func TestChatWrite_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	db := setupChatTestDB(t)
	defer db.Close()

	repo := NewChatWriteRepository(db)

	chatUUID := uuid.New().String()
	createdByUUID := uuid.New().String()
	participants := createdByUUID + ",user2"

	err := repo.Save(ctx, chatUUID, createdByUUID, participants)
	assert.NoError(t, err)

	var oldUpdatedAt time.Time
	err = db.Get(&oldUpdatedAt, "SELECT updated_at FROM chats WHERE chat_uuid = ?", chatUUID)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second) // чтобы updated_at изменился

	newParticipants := participants + ",extra-user"
	err = repo.Save(ctx, chatUUID, createdByUUID, newParticipants)
	assert.NoError(t, err)

	var got models.ChatDB
	err = db.Get(&got, `
		SELECT chat_uuid, participants_uuids, created_by_uuid, created_at, updated_at
		FROM chats WHERE chat_uuid = ?`, chatUUID)
	assert.NoError(t, err)
	assert.Equal(t, newParticipants, got.ParticipantsUUIDs)
	assert.True(t, got.UpdatedAt.After(oldUpdatedAt))
}

func TestChatRead_Get(t *testing.T) {
	ctx := context.Background()
	db := setupChatTestDB(t)
	defer db.Close()

	writeRepo := NewChatWriteRepository(db)
	readRepo := NewChatReadRepository(db)

	chatUUID := uuid.New().String()
	createdByUUID := uuid.New().String()
	participants := createdByUUID + ",user2,user3"

	err := writeRepo.Save(ctx, chatUUID, createdByUUID, participants)
	assert.NoError(t, err)

	got, err := readRepo.Get(ctx, chatUUID)
	assert.NoError(t, err)
	assert.Equal(t, chatUUID, got.ChatUUID)
	assert.Equal(t, createdByUUID, got.CreatedByUUID)
	assert.Equal(t, participants, got.ParticipantsUUIDs)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
}

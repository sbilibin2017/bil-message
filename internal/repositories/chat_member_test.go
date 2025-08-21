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

func setupChatMemberTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect to sqlite: %v", err)
	}

	schema := `
	CREATE TABLE chat_members (
		chat_member_uuid TEXT PRIMARY KEY,
		chat_uuid TEXT NOT NULL,
		user_uuid TEXT NOT NULL,
		joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(chat_uuid, user_uuid)
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create chat_members table: %v", err)
	}

	return db
}

func TestChatMemberWrite_Save(t *testing.T) {
	ctx := context.Background()
	db := setupChatMemberTestDB(t)
	defer db.Close()

	repo := NewChatMemberWriteRepository(db)

	chatMemberUUID := uuid.New()
	chatUUID := uuid.New()
	userUUID := uuid.New()

	member := &models.ChatMemberDB{
		ChatMemberUUID: chatMemberUUID.String(),
		ChatUUID:       chatUUID.String(),
		UserUUID:       userUUID.String(),
	}

	// Сохраняем нового участника
	err := repo.Save(ctx, member)
	assert.NoError(t, err)

	// Проверяем запись
	var got models.ChatMemberDB
	err = db.Get(&got, "SELECT chat_member_uuid, chat_uuid, user_uuid, joined_at, created_at, updated_at FROM chat_members WHERE chat_member_uuid = ?", chatMemberUUID.String())
	assert.NoError(t, err)
	assert.Equal(t, chatMemberUUID.String(), got.ChatMemberUUID)
	assert.Equal(t, chatUUID.String(), got.ChatUUID)
	assert.Equal(t, userUUID.String(), got.UserUUID)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
}

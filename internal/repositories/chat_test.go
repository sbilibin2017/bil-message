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

	chatUUID := uuid.New()
	createdByUUID := uuid.New()

	chat := &models.ChatDB{
		ChatUUID:      chatUUID.String(),
		CreatedByUUID: createdByUUID.String(),
	}

	// Сохраняем новый чат
	err := repo.Save(ctx, chat)
	assert.NoError(t, err)

	// Проверяем запись в БД
	var got models.ChatDB
	err = db.Get(&got, "SELECT chat_uuid, created_by_uuid, created_at, updated_at FROM chats WHERE chat_uuid = ?", chatUUID.String())
	assert.NoError(t, err)
	assert.Equal(t, chatUUID.String(), got.ChatUUID)
	assert.Equal(t, createdByUUID.String(), got.CreatedByUUID)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
}

func TestChatWrite_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	db := setupChatTestDB(t)
	defer db.Close()

	repo := NewChatWriteRepository(db)

	chatUUID := uuid.New()
	createdByUUID := uuid.New()

	chat := &models.ChatDB{
		ChatUUID:      chatUUID.String(),
		CreatedByUUID: createdByUUID.String(),
	}

	// Сохраняем первый раз
	err := repo.Save(ctx, chat)
	assert.NoError(t, err)

	// Получаем старое updated_at
	var oldUpdatedAt time.Time
	err = db.Get(&oldUpdatedAt, "SELECT updated_at FROM chats WHERE chat_uuid = ?", chatUUID.String())
	assert.NoError(t, err)

	// Ждём немного, чтобы updated_at точно изменился
	time.Sleep(1 * time.Second)

	// Сохраняем повторно (обновление)
	err = repo.Save(ctx, chat)
	assert.NoError(t, err)

	// Проверяем, что updated_at изменился
	var newUpdatedAt time.Time
	err = db.Get(&newUpdatedAt, "SELECT updated_at FROM chats WHERE chat_uuid = ?", chatUUID.String())
	assert.NoError(t, err)
	assert.True(t, newUpdatedAt.After(oldUpdatedAt), "updated_at должен обновиться")
}

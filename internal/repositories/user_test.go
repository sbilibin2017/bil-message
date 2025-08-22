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

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	assert.NoError(t, err)

	schema := `
	CREATE TABLE users (
		user_uuid TEXT PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		public_key TEXT,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	assert.NoError(t, err)
	return db
}

func TestUserWriteAndRead(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	writeRepo := NewUserWriteRepository(db)
	readRepo := NewUserReadRepository(db)

	userUUID := uuid.New().String()
	username := "testuser"
	passwordHash := "hash123"

	// Сохраняем пользователя
	err := writeRepo.Save(ctx, userUUID, username, passwordHash)
	assert.NoError(t, err)

	// Читаем по username
	got, err := readRepo.Get(ctx, username)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, userUUID, got.UserUUID)
	assert.Equal(t, username, got.Username)
	assert.Equal(t, passwordHash, got.PasswordHash)
	assert.WithinDuration(t, time.Now(), got.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
}

func TestUserWrite_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	writeRepo := NewUserWriteRepository(db)
	readRepo := NewUserReadRepository(db)

	userUUID := uuid.New().String()
	username := "existinguser"
	passwordHash1 := "hash1"
	passwordHash2 := "hash2"

	// Первый сохраненный пользователь
	err := writeRepo.Save(ctx, userUUID, username, passwordHash1)
	assert.NoError(t, err)

	// Обновляем пароль
	err = writeRepo.Save(ctx, userUUID, username, passwordHash2)
	assert.NoError(t, err)

	got, err := readRepo.Get(ctx, username)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, passwordHash2, got.PasswordHash)
	assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Second)
}

func TestUserRead_NotFound(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	readRepo := NewUserReadRepository(db)
	got, err := readRepo.Get(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestUserRead_GetByUUID(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	writeRepo := NewUserWriteRepository(db)

	userUUID := uuid.New().String()
	username := "uuiduser"
	passwordHash := "uuidhash"

	err := writeRepo.Save(ctx, userUUID, username, passwordHash)
	assert.NoError(t, err)

}

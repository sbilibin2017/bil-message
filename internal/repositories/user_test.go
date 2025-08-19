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

// setupTestDB создает таблицу users с полями CreatedAt и UpdatedAt.
func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect to sqlite: %v", err)
	}

	schema := `
	CREATE TABLE users (
		user_uuid TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	return db
}

func TestUserWriteAndRead(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	writeRepo := NewUserWriteRepository(db)
	readRepo := NewUserReadRepository(db)

	userUUID := uuid.New()
	username := "testuser"
	passwordHash := "hash123"

	// Save user
	err := writeRepo.Save(ctx, userUUID, username, passwordHash)
	assert.NoError(t, err)

	// Get user
	user, err := readRepo.Get(ctx, username)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, userUUID, user.UserUUID)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, passwordHash, user.PasswordHash)

	// Проверяем, что CreatedAt и UpdatedAt проставлены
	assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)
}

func TestUserWrite_UpdateExisting(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	writeRepo := NewUserWriteRepository(db)
	readRepo := NewUserReadRepository(db)

	userUUID := uuid.New()
	username := "existinguser"
	passwordHash1 := "hash1"
	passwordHash2 := "hash2"

	// First save
	err := writeRepo.Save(ctx, userUUID, username, passwordHash1)
	assert.NoError(t, err)

	// Update existing user
	err = writeRepo.Save(ctx, userUUID, username, passwordHash2)
	assert.NoError(t, err)

	user, err := readRepo.Get(ctx, username)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, passwordHash2, user.PasswordHash)
}

func TestUserRead_NotFound(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	readRepo := NewUserReadRepository(db)

	user, err := readRepo.Get(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, user) // теперь проверяем, что пользователь не найден и возвращается nil
}

func TestUserRead_GetError(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	readRepo := NewUserReadRepository(db)

	// Intentionally break the table to simulate a SQL error
	_, err := db.Exec(`DROP TABLE users`)
	assert.NoError(t, err)

	user, err := readRepo.Get(ctx, "anyuser")
	assert.Error(t, err) // Should return an actual SQL error
	assert.Nil(t, user)  // User should be nil on error
}

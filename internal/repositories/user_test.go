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

// setupTestDB создаёт in-memory SQLite БД с таблицей users
func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to connect to sqlite: %v", err)
	}

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

	userUUID := uuid.New().String()
	username := "testuser"
	passwordHash := "hash123"

	user := &models.UserDB{
		UserUUID:     userUUID,
		Username:     username,
		PasswordHash: passwordHash,
	}

	// Save user
	err := writeRepo.Save(ctx, user)
	assert.NoError(t, err)

	// Get user
	got, err := readRepo.GetByUsername(ctx, username)
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

	user := &models.UserDB{
		UserUUID:     userUUID,
		Username:     username,
		PasswordHash: passwordHash1,
	}

	// First save
	err := writeRepo.Save(ctx, user)
	assert.NoError(t, err)

	// Update existing user
	user.PasswordHash = passwordHash2
	err = writeRepo.Save(ctx, user)
	assert.NoError(t, err)

	got, err := readRepo.GetByUsername(ctx, username)
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

	got, err := readRepo.GetByUsername(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestUserRead_GetByUUID(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	writeRepo := NewUserWriteRepository(db)
	readRepo := NewUserReadRepository(db)

	userUUID := uuid.New().String()
	username := "uuiduser"
	passwordHash := "uuidhash"

	user := &models.UserDB{
		UserUUID:     userUUID,
		Username:     username,
		PasswordHash: passwordHash,
	}

	err := writeRepo.Save(ctx, user)
	assert.NoError(t, err)

	got, err := readRepo.GetByUUID(ctx, userUUID)
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, userUUID, got.UserUUID)
	assert.Equal(t, username, got.Username)
	assert.Equal(t, passwordHash, got.PasswordHash)
}

func TestUserRead_GetByUUID_NotFound(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	readRepo := NewUserReadRepository(db)

	randomUUID := uuid.New().String()
	got, err := readRepo.GetByUUID(ctx, randomUUID)
	assert.NoError(t, err)
	assert.Nil(t, got)
}

func TestUserRead_GetByUUID_Error(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	readRepo := NewUserReadRepository(db)

	// ломаем таблицу
	_, err := db.Exec(`DROP TABLE users`)
	assert.NoError(t, err)

	randomUUID := uuid.New().String()
	got, err := readRepo.GetByUUID(ctx, randomUUID)
	assert.Error(t, err)
	assert.Nil(t, got)
}

package repositories_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/repositories"
	"github.com/stretchr/testify/assert"

	_ "modernc.org/sqlite" // sqlite driver
)

func setupDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Open("sqlite", ":memory:")
	assert.NoError(t, err)

	schema := `
	CREATE TABLE users (
		user_uuid     TEXT PRIMARY KEY,
		username      TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at    DATETIME NOT NULL,
		updated_at    DATETIME NOT NULL
	);`
	_, err = db.Exec(schema)
	assert.NoError(t, err)

	return db
}

func TestUserWriteAndRead(t *testing.T) {
	db := setupDB(t)
	writeRepo := repositories.NewUserWriteRepository(db)
	readRepo := repositories.NewUserReadRepository(db)
	ctx := context.Background()

	userUUID := uuid.New()
	username := "johndoe"
	passwordHash := "hashedpass"

	// Save new user
	err := writeRepo.Save(ctx, userUUID, username, passwordHash)
	assert.NoError(t, err)

	// Get by username
	user, err := readRepo.Get(ctx, username)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, username, user.Username)
	assert.Equal(t, passwordHash, user.PasswordHash)
	assert.Equal(t, userUUID, user.UserUUID)

	// Save with same user_uuid but updated info
	newUsername := "johnupdated"
	newPasswordHash := "newhash"
	err = writeRepo.Save(ctx, userUUID, newUsername, newPasswordHash)
	assert.NoError(t, err)

	// Read again
	user, err = readRepo.Get(ctx, newUsername)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, newUsername, user.Username)
	assert.Equal(t, newPasswordHash, user.PasswordHash)
	assert.Equal(t, userUUID, user.UserUUID)

	// Check old username does not exist
	user, err = readRepo.Get(ctx, username)
	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestGetNonExistingUser(t *testing.T) {
	db := setupDB(t)
	readRepo := repositories.NewUserReadRepository(db)
	ctx := context.Background()

	user, err := readRepo.Get(ctx, "nonexistent")
	assert.NoError(t, err)
	assert.Nil(t, user)
}

func TestSaveMultipleUsers(t *testing.T) {
	db := setupDB(t)
	writeRepo := repositories.NewUserWriteRepository(db)
	readRepo := repositories.NewUserReadRepository(db)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		id := uuid.New()
		username := "user" + string(rune(i+'0'))
		password := "pass" + string(rune(i+'0'))
		err := writeRepo.Save(ctx, id, username, password)
		assert.NoError(t, err)

		user, err := readRepo.Get(ctx, username)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, username, user.Username)
	}
}

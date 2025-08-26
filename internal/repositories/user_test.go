package repositories_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/repositories"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	db, err := sqlx.Connect("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	// Создаём таблицу users
	schema := `
	CREATE TABLE users (
		user_uuid TEXT PRIMARY KEY,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatal(err)
	}

	return db
}

func TestUserRepositories(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	userWriteRepo := repositories.NewUserWriteRepository(db)
	userReadRepo := repositories.NewUserReadRepository(db)

	userUUID := uuid.New()
	username := "alice"
	password := "password123"

	// --- Test Save ---
	err := userWriteRepo.Save(ctx, userUUID, username, password)
	assert.NoError(t, err, "expected no error saving user")

	// --- Test GetByUsername ---
	userByName, err := userReadRepo.Get(ctx, username)
	assert.NoError(t, err)
	assert.NotNil(t, userByName)
	assert.Equal(t, username, userByName.Username)
	assert.Equal(t, password, userByName.Password)
	assert.Equal(t, userUUID, userByName.UserUUID)

	// --- Test Update existing user ---
	newPassword := "newpass456"
	err = userWriteRepo.Save(ctx, userUUID, username, newPassword)
	assert.NoError(t, err)

	updatedUser, err := userReadRepo.Get(ctx, username)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.Equal(t, newPassword, updatedUser.Password)

	// --- Test Get non-existent user ---
	nonUser, err := userReadRepo.Get(ctx, "bob")
	assert.NoError(t, err)
	assert.Nil(t, nonUser)
}

func TestUserWriteRepository_Conflict_Update(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)
	defer db.Close()

	userWriteRepo := repositories.NewUserWriteRepository(db)
	userReadRepo := repositories.NewUserReadRepository(db)
	userUUID := uuid.New()

	// Insert first user
	err := userWriteRepo.Save(ctx, userUUID, "alice", "pass1")
	assert.NoError(t, err)

	// Insert same user_uuid but different username and password -> should update
	err = userWriteRepo.Save(ctx, userUUID, "alice2", "pass2")
	assert.NoError(t, err)

	user, err := userReadRepo.Get(ctx, "alice2")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "pass2", user.Password)
	assert.Equal(t, userUUID, user.UserUUID)
}

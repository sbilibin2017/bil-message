package repositories

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// UserWriteRepository реализует интерфейс UserWriter через sqlx.DB.
type UserWriteRepository struct {
	db *sqlx.DB
}

func NewUserWriteRepository(db *sqlx.DB) *UserWriteRepository {
	return &UserWriteRepository{db: db}
}

// Save сохраняет пользователя в базе данных.
func (r *UserWriteRepository) Save(
	ctx context.Context,
	user *models.UserDB,
) error {
	query := `
		INSERT INTO users (user_uuid, username, password_hash)
		VALUES ($1, $2, $3)
		ON CONFLICT(username) DO UPDATE
		SET password_hash = excluded.password_hash,
		    updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query, user.UserUUID, user.Username, user.PasswordHash)
	return err
}

// UserReadRepository реализует интерфейс UserReader через sqlx.DB.
type UserReadRepository struct {
	db *sqlx.DB
}

func NewUserReadRepository(db *sqlx.DB) *UserReadRepository {
	return &UserReadRepository{db: db}
}

// GetByUsername возвращает пользователя по username.
func (r *UserReadRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*models.UserDB, error) {
	var user models.UserDB
	query := `SELECT user_uuid, username, password_hash, created_at, updated_at FROM users WHERE username = $1`
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByUUID возвращает пользователя по userUUID.
func (r *UserReadRepository) GetByUUID(
	ctx context.Context,
	userUUID string,
) (*models.UserDB, error) {
	var user models.UserDB
	query := `SELECT user_uuid, username, password_hash, created_at, updated_at 
	          FROM users 
	          WHERE user_uuid = $1`

	err := r.db.GetContext(ctx, &user, query, userUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// UserWriteRepository репозиторий для записи пользователей в базу
type UserWriteRepository struct {
	db *sqlx.DB
}

// NewUserWriteRepository создаёт новый репозиторий для записи пользователей
func NewUserWriteRepository(db *sqlx.DB) *UserWriteRepository {
	return &UserWriteRepository{db: db}
}

// Save сохраняет нового пользователя или обновляет существующего по user_uuid
func (r *UserWriteRepository) Save(
	ctx context.Context,
	userUUID uuid.UUID,
	username string,
	password string,
) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (user_uuid, username, password, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_uuid)
		 DO UPDATE
		 SET username = EXCLUDED.username,
		     password = EXCLUDED.password,
		     updated_at = EXCLUDED.updated_at`,
		userUUID, username, password, now, now,
	)
	return err
}

// UserReadRepository репозиторий для чтения пользователей из базы
type UserReadRepository struct {
	db *sqlx.DB
}

// NewUserReadRepository создаёт новый репозиторий для чтения пользователей
func NewUserReadRepository(db *sqlx.DB) *UserReadRepository {
	return &UserReadRepository{db: db}
}

// GetByUsername возвращает пользователя по username или nil, если не найден
func (r *UserReadRepository) GetByUsername(
	ctx context.Context,
	username string,
) (*models.UserDB, error) {
	var user models.UserDB
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE username = $1", username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByUUID возвращает пользователя по user_uuid или nil, если не найден
func (r *UserReadRepository) GetByUUID(
	ctx context.Context,
	userUUID uuid.UUID,
) (*models.UserDB, error) {
	var user models.UserDB
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE user_uuid = $1", userUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

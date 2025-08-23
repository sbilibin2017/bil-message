package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// UserWriteRepository реализует интерфейс UserSaver через SQL базу
type UserWriteRepository struct {
	db *sqlx.DB
}

// NewUserWriteRepository создаёт новый репозиторий для записи пользователей
func NewUserWriteRepository(db *sqlx.DB) *UserWriteRepository {
	return &UserWriteRepository{db: db}
}

// Save сохраняет нового пользователя в базу.
// Если пользователь с таким user_uuid уже существует, обновляются только username, password_hash и updated_at.
// user_uuid остаётся неизменным.
func (r *UserWriteRepository) Save(
	ctx context.Context,
	userUUID uuid.UUID,
	username string,
	passwordHash string,
) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (user_uuid, username, password_hash, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (user_uuid)
		 DO UPDATE
		 SET username = EXCLUDED.username,
		     password_hash = EXCLUDED.password_hash,
		     updated_at = EXCLUDED.updated_at`,
		userUUID, username, passwordHash, now, now,
	)
	return err
}

// UserReadRepository реализует интерфейс UserGetter через SQL базу
type UserReadRepository struct {
	db *sqlx.DB
}

// NewUserReadRepository создаёт новый репозиторий для чтения пользователей
func NewUserReadRepository(db *sqlx.DB) *UserReadRepository {
	return &UserReadRepository{db: db}
}

// Get возвращает пользователя по username или nil, если не найден
func (r *UserReadRepository) Get(
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

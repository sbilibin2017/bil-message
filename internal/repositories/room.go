package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// RoomWriteRepository реализует интерфейс сохранения комнат через SQL базу
type RoomWriteRepository struct {
	db *sqlx.DB
}

// NewRoomWriteRepository создаёт новый репозиторий для записи комнат
func NewRoomWriteRepository(db *sqlx.DB) *RoomWriteRepository {
	return &RoomWriteRepository{db: db}
}

// Save сохраняет новую комнату в базу.
// Если комната с таким room_uuid уже существует, обновляется только updated_at.
func (r *RoomWriteRepository) Save(
	ctx context.Context,
	roomUUID uuid.UUID,
	creatorUUID uuid.UUID,
) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO rooms (room_uuid, creator_uuid, created_at, updated_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (room_uuid)
		 DO UPDATE
		 SET updated_at = EXCLUDED.updated_at`,
		roomUUID, creatorUUID, now, now,
	)
	if err != nil {
		return err
	}
	return err
}

// Delete удаляет комнату из базы по roomUUID
func (r *RoomWriteRepository) Delete(ctx context.Context, roomUUID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM rooms WHERE room_uuid = $1`,
		roomUUID,
	)
	return err
}

// RoomReadRepository реализует интерфейс получения комнат через SQL базу
type RoomReadRepository struct {
	db *sqlx.DB
}

// NewRoomReadRepository создаёт новый репозиторий для чтения комнат
func NewRoomReadRepository(db *sqlx.DB) *RoomReadRepository {
	return &RoomReadRepository{db: db}
}

// Get возвращает комнату по UUID или nil, если не найдена
func (r *RoomReadRepository) Get(
	ctx context.Context,
	roomUUID uuid.UUID,
) (*models.RoomDB, error) {
	var room models.RoomDB
	err := r.db.GetContext(ctx, &room, "SELECT * FROM rooms WHERE room_uuid = $1", roomUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &room, nil
}

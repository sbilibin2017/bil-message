package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// RoomWriteRepository репозиторий для записи комнат
type RoomWriteRepository struct {
	db *sqlx.DB
}

func NewRoomWriteRepository(db *sqlx.DB) *RoomWriteRepository {
	return &RoomWriteRepository{db: db}
}

// Save сохраняет новую комнату или обновляет существующую по room_uuid
func (r *RoomWriteRepository) Save(ctx context.Context, roomUUID, ownerUUID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO rooms (room_uuid, owner_uuid, created_at, updated_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (room_uuid)
		 DO UPDATE
		 SET owner_uuid = EXCLUDED.owner_uuid,
		     updated_at = EXCLUDED.updated_at`,
		roomUUID, ownerUUID, now, now,
	)
	return err
}

// Delete удаляет комнату по room_uuid
func (r *RoomWriteRepository) Delete(ctx context.Context, roomUUID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM rooms WHERE room_uuid = $1", roomUUID)
	return err
}

// RoomReadRepository репозиторий для чтения комнат
type RoomReadRepository struct {
	db *sqlx.DB
}

func NewRoomReadRepository(db *sqlx.DB) *RoomReadRepository {
	return &RoomReadRepository{db: db}
}

// Get возвращает комнату по room_uuid или nil, если не найдена
func (r *RoomReadRepository) Get(ctx context.Context, roomUUID uuid.UUID) (*models.RoomDB, error) {
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

package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// RoomMemberWriteRepository репозиторий для записи участников комнат
type RoomMemberWriteRepository struct {
	db *sqlx.DB
}

func NewRoomMemberWriteRepository(db *sqlx.DB) *RoomMemberWriteRepository {
	return &RoomMemberWriteRepository{db: db}
}

// Save сохраняет нового участника комнаты
func (r *RoomMemberWriteRepository) Save(ctx context.Context, roomUUID, memberUUID uuid.UUID) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO room_members (room_uuid, member_uuid, created_at, updated_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (room_uuid, member_uuid)
		 DO UPDATE
		 SET updated_at = EXCLUDED.updated_at`,
		roomUUID, memberUUID, now, now,
	)
	return err
}

// Delete удаляет участника комнаты
func (r *RoomMemberWriteRepository) Delete(ctx context.Context, roomUUID, memberUUID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM room_members WHERE room_uuid = $1 AND member_uuid = $2",
		roomUUID, memberUUID,
	)
	return err
}

// RoomMemberReadRepository репозиторий для чтения участников комнат
type RoomMemberReadRepository struct {
	db *sqlx.DB
}

func NewRoomMemberReadRepository(db *sqlx.DB) *RoomMemberReadRepository {
	return &RoomMemberReadRepository{db: db}
}

// Get возвращает участника комнаты по room_uuid и user_uuid
func (r *RoomMemberReadRepository) Get(ctx context.Context, roomUUID, memberUUID uuid.UUID) (*models.RoomMemberDB, error) {
	var member models.RoomMemberDB
	err := r.db.GetContext(ctx, &member,
		"SELECT * FROM room_members WHERE room_uuid = $1 AND member_uuid = $2",
		roomUUID, memberUUID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

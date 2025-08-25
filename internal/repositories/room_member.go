package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// RoomMemberWriteRepository реализует запись участников комнаты через SQL
type RoomMemberWriteRepository struct {
	db *sqlx.DB
}

// NewRoomMemberWriteRepository создаёт новый репозиторий для участников комнат
func NewRoomMemberWriteRepository(db *sqlx.DB) *RoomMemberWriteRepository {
	return &RoomMemberWriteRepository{db: db}
}

// Save добавляет пользователя в комнату.
// Если запись уже существует, обновляется только updated_at.
func (r *RoomMemberWriteRepository) Save(
	ctx context.Context,
	roomUUID uuid.UUID,
	userUUID uuid.UUID,
	joinedAt time.Time,
) error {
	now := time.Now().UTC()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO room_members (room_uuid, user_uuid, joined_at, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (room_uuid, user_uuid)
		 DO UPDATE
		 SET updated_at = EXCLUDED.updated_at`,
		roomUUID, userUUID, joinedAt, now, now,
	)
	return err
}

// Delete удаляет конкретного пользователя из комнаты
func (r *RoomMemberWriteRepository) Delete(ctx context.Context, roomUUID uuid.UUID, userUUID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM room_members WHERE room_uuid = $1 AND user_uuid = $2`,
		roomUUID, userUUID,
	)
	return err
}

// RoomMemberReadRepository реализует чтение участников комнаты через SQL
type RoomMemberReadRepository struct {
	db *sqlx.DB
}

// NewRoomMemberReadRepository создаёт новый репозиторий для чтения участников
func NewRoomMemberReadRepository(db *sqlx.DB) *RoomMemberReadRepository {
	return &RoomMemberReadRepository{db: db}
}

// GetAllByRoom возвращает всех участников комнаты
func (r *RoomMemberReadRepository) List(
	ctx context.Context,
	roomUUID uuid.UUID,
) ([]models.RoomMemberDB, error) {
	var members []models.RoomMemberDB
	err := r.db.SelectContext(ctx, &members,
		`SELECT * FROM room_members WHERE room_uuid = $1`, roomUUID)
	return members, err
}

// Get возвращает участника комнаты по roomUUID и userUUID
func (r *RoomMemberReadRepository) Get(
	ctx context.Context,
	roomUUID, userUUID uuid.UUID,
) (*models.RoomMemberDB, error) {
	var member models.RoomMemberDB
	err := r.db.GetContext(ctx, &member,
		`SELECT * FROM room_members WHERE room_uuid = $1 AND user_uuid = $2`,
		roomUUID, userUUID,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return &member, nil
}

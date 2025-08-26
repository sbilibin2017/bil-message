package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/models"
)

type DeviceWriteRepository struct {
	db *sqlx.DB
}

func NewDeviceWriteRepository(db *sqlx.DB) *DeviceWriteRepository {
	return &DeviceWriteRepository{db: db}
}

func (r *DeviceWriteRepository) Save(
	ctx context.Context,
	deviceUUID uuid.UUID,
	userUUID uuid.UUID,
	publicKey string,
) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO devices (device_uuid, user_uuid, public_key, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (device_uuid)
		 DO UPDATE
		 SET user_uuid = EXCLUDED.user_uuid,
		     public_key = EXCLUDED.public_key,
		     updated_at = EXCLUDED.updated_at`,
		deviceUUID, userUUID, publicKey, now, now,
	)
	return err
}

type DeviceReadRepository struct {
	db *sqlx.DB
}

func NewDeviceReadRepository(db *sqlx.DB) *DeviceReadRepository {
	return &DeviceReadRepository{db: db}
}

func (r *DeviceReadRepository) Get(
	ctx context.Context,
	deviceUUID uuid.UUID,
) (*models.DeviceDB, error) {
	var device models.DeviceDB
	err := r.db.GetContext(ctx, &device, "SELECT * FROM devices WHERE device_uuid = $1", deviceUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

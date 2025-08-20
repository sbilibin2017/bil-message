package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// DeviceWriteRepository реализует интерфейс ClientWriter через sqlx.DB.
type DeviceWriteRepository struct {
	db *sqlx.DB
}

// NewDeviceWriteRepository создаёт новый репозиторий для работы с клиентами.
func NewDeviceWriteRepository(db *sqlx.DB) *DeviceWriteRepository {
	return &DeviceWriteRepository{db: db}
}

// Save сохраняет клиента в базе данных.
func (r *DeviceWriteRepository) Save(
	ctx context.Context,
	deviceUUID uuid.UUID,
	userUUID uuid.UUID,
	publicKey string,
) error {
	query := `
	INSERT INTO devices (device_uuid, user_uuid, public_key)
	VALUES ($1, $2, $3)
	ON CONFLICT(device_uuid) DO UPDATE
	SET public_key = excluded.public_key,
	    updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query, deviceUUID, userUUID, publicKey)
	return err
}

// DeviceReadRepository реализует чтение клиентов через sqlx.DB.
type DeviceReadRepository struct {
	db *sqlx.DB
}

// NewDeviceReadRepository создаёт новый репозиторий для работы с клиентами.
func NewDeviceReadRepository(db *sqlx.DB) *DeviceReadRepository {
	return &DeviceReadRepository{db: db}
}

// GetByUUID возвращает устройство по его UUID.
func (r *DeviceReadRepository) GetByUUID(
	ctx context.Context,
	deviceUUID uuid.UUID,
) (*models.DeviceDB, error) {
	var device models.DeviceDB
	query := `SELECT device_uuid, user_uuid, public_key, created_at, updated_at 
	          FROM devices WHERE device_uuid = $1`
	err := r.db.GetContext(ctx, &device, query, deviceUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

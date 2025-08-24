package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// DeviceWriteRepository реализует интерфейс DeviceSaver через SQL базу
type DeviceWriteRepository struct {
	db *sqlx.DB
}

// NewDeviceWriteRepository создаёт новый репозиторий для записи устройств
func NewDeviceWriteRepository(db *sqlx.DB) *DeviceWriteRepository {
	return &DeviceWriteRepository{db: db}
}

// Save сохраняет новое устройство с привязкой к пользователю.
// Если устройство с таким device_uuid уже существует, обновляются только public_key и updated_at.
func (r *DeviceWriteRepository) Save(
	ctx context.Context,
	deviceUUID uuid.UUID,
	userUUID uuid.UUID,
	publicKey string,
) error {
	now := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_devices (device_uuid, user_uuid, public_key, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (device_uuid)
		 DO UPDATE
		 SET public_key = EXCLUDED.public_key,
		     updated_at = EXCLUDED.updated_at`,
		deviceUUID, userUUID, publicKey, now, now,
	)
	return err
}

// DeviceReadRepository реализует чтение устройств пользователя через SQL базу
type DeviceReadRepository struct {
	db *sqlx.DB
}

// NewDeviceReadRepository создаёт новый репозиторий для чтения устройств
func NewDeviceReadRepository(db *sqlx.DB) *DeviceReadRepository {
	return &DeviceReadRepository{db: db}
}

// GetByDeviceUUID возвращает устройство по его UUID или nil, если не найдено
func (r *DeviceReadRepository) Get(
	ctx context.Context,
	deviceUUID uuid.UUID,
) (*models.UserDeviceDB, error) {
	var device models.UserDeviceDB
	err := r.db.GetContext(ctx, &device, "SELECT * FROM user_devices WHERE device_uuid = $1", deviceUUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

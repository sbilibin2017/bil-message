package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/errors"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// DeviceUserReader определяет интерфейс для чтения данных пользователя из хранилища.
type DeviceUserReader interface {
	// GetByUUID возвращает пользователя по userUUID.
	GetByUUID(ctx context.Context, userUUID uuid.UUID) (*models.UserDB, error)
}

// DeviceWriter определяет интерфейс для сохранения данных устройства.
type DeviceWriter interface {
	Save(ctx context.Context, deviceUUID uuid.UUID, userUUID uuid.UUID, publicKey string) error
}

// DeviceService управляет устройствами пользователя.
type DeviceService struct {
	ur DeviceUserReader
	dw DeviceWriter
}

// NewDeviceService создаёт новый экземпляр DeviceService.
func NewDeviceService(
	ur DeviceUserReader,
	dw DeviceWriter,
) *DeviceService {
	return &DeviceService{
		ur: ur,
		dw: dw,
	}
}

// Register регистрирует новое устройство пользователя.
func (svc *DeviceService) Register(
	ctx context.Context,
	userUUID uuid.UUID,
	publicKey string,
) (*uuid.UUID, error) {
	// 1. Проверяем, что пользователь существует
	user, err := svc.ur.GetByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	// 2. Генерируем UUID устройства
	deviceUUID := uuid.New()

	// 3. Сохраняем устройство
	if err := svc.dw.Save(ctx, deviceUUID, userUUID, publicKey); err != nil {
		return nil, err
	}

	// 4. Возвращаем UUID устройства
	return &deviceUUID, nil
}

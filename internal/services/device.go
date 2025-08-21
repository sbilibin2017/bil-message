package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/errors"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// DeviceUserReader определяет интерфейс для чтения данных пользователя из хранилища.
type DeviceUserReader interface {
	GetByUUID(ctx context.Context, userUUID string) (*models.UserDB, error)
}

// DeviceWriter определяет интерфейс для сохранения данных устройства.
type DeviceWriter interface {
	Save(ctx context.Context, device *models.DeviceDB) error
}

// DeviceService управляет устройствами пользователя.
type DeviceService struct {
	ur DeviceUserReader
	dw DeviceWriter
}

// NewDeviceService создаёт новый экземпляр DeviceService.
func NewDeviceService(ur DeviceUserReader, dw DeviceWriter) *DeviceService {
	return &DeviceService{
		ur: ur,
		dw: dw,
	}
}

// Register регистрирует новое устройство пользователя.
func (svc *DeviceService) Register(ctx context.Context, userUUID string, publicKey string) (*string, error) {
	// Проверяем, что пользователь существует
	user, err := svc.ur.GetByUUID(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	// Генерируем UUID устройства
	deviceUUID := uuid.New().String()

	// Создаём DeviceDB
	device := &models.DeviceDB{
		DeviceUUID: deviceUUID,
		UserUUID:   userUUID,
		PublicKey:  publicKey,
	}

	// Сохраняем устройство
	if err := svc.dw.Save(ctx, device); err != nil {
		return nil, err
	}

	return &deviceUUID, nil
}

package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/errors"
	"github.com/sbilibin2017/bil-message/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserReader определяет интерфейс для чтения данных пользователя из хранилища.
type UserReader interface {
	GetByUsername(ctx context.Context, username string) (*models.UserDB, error)
	GetByUUID(ctx context.Context, userUUID string) (*models.UserDB, error)
}

// UserWriter определяет интерфейс для сохранения данных пользователя в хранилище.
type UserWriter interface {
	Save(ctx context.Context, user *models.UserDB) error
}

// DeviceReader
type DeviceReader interface {
	GetByUUID(ctx context.Context, deviceUUID string) (*models.DeviceDB, error)
}

// TokenGenerator
type TokenGenerator interface {
	Generate(payload *models.TokenPayload) (string, error)
}

// AuthService предоставляет методы для создания пользователей и клиентов.
type AuthService struct {
	ur UserReader
	uw UserWriter
	dr DeviceReader
	tg TokenGenerator
}

func NewAuthService(
	ur UserReader,
	uw UserWriter,
	dr DeviceReader,
	tg TokenGenerator,
) *AuthService {
	return &AuthService{
		ur: ur,
		uw: uw,
		dr: dr,
		tg: tg,
	}
}

// Register создаёт нового пользователя и возвращает его UUID.
func (svc *AuthService) Register(ctx context.Context, username, password string) (userUUID string, err error) {
	// Проверяем, существует ли пользователь
	existingUser, err := svc.ur.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if existingUser != nil {
		return "", errors.ErrUserAlreadyExists
	}

	// Хэшируем пароль
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// Создаём пользователя
	user := &models.UserDB{
		UserUUID:     uuid.New().String(),
		Username:     username,
		PasswordHash: string(hashBytes),
	}

	// Сохраняем
	if err := svc.uw.Save(ctx, user); err != nil {
		return "", err
	}

	return user.UserUUID, nil
}

// Login проверяет пользователя и устройство, затем возвращает токен.
func (svc *AuthService) Login(ctx context.Context, username, password string, deviceUUID string) (token string, err error) {
	// Получаем пользователя
	user, err := svc.ur.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.ErrUserNotFound
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.ErrInvalidPassword
	}

	// Проверяем устройство
	device, err := svc.dr.GetByUUID(ctx, deviceUUID)
	if err != nil {
		return "", err
	}
	if device == nil || device.UserUUID != user.UserUUID {
		return "", errors.ErrDeviceNotFound
	}

	payload := models.TokenPayload{
		UserUUID:   user.UserUUID,
		DeviceUUID: device.DeviceUUID,
	}
	// Генерируем токен
	token, err = svc.tg.Generate(&payload)
	if err != nil {
		return "", err
	}

	return token, nil
}

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
	// GetByUsername возвращает пользователя по username.
	GetByUsername(ctx context.Context, username string) (*models.UserDB, error)
}

// UserWriter определяет интерфейс для сохранения данных пользователя в хранилище.
type UserWriter interface {
	// Save сохраняет пользователя с заданным UUID, именем и хэшем пароля.
	Save(ctx context.Context, userUUID uuid.UUID, username string, passwordHash string) error
}

// DeviceReader
type DeviceReader interface {
	GetByUUID(ctx context.Context, deviceUUID uuid.UUID) (*models.DeviceDB, error)
}

type TokenGenerator interface {
	Generate(userUUID uuid.UUID, clientUUID uuid.UUID) (string, error)
}

// AuthService предоставляет методы для создания пользователей и клиентов.
type AuthService struct {
	ur UserReader
	uw UserWriter
	dr DeviceReader
	tg TokenGenerator
}

// NewAuthService создаёт новый экземпляр AuthService.
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

// Register создаёт нового пользователя, клиента с RSA ключами и возвращает токен и приватный ключ.
func (svc *AuthService) Register(
	ctx context.Context,
	username string,
	password string,
) (userUUID uuid.UUID, err error) {
	// 1. Проверяем, существует ли пользователь
	existingUser, err := svc.ur.GetByUsername(ctx, username)
	if err != nil {
		return uuid.Nil, err
	}
	if existingUser != nil {
		return uuid.Nil, errors.ErrUserAlreadyExists
	}

	// 2. Хэшируем пароль
	passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, err
	}
	passwordHash := string(passwordHashBytes)

	// 3. Генерируем UUID пользователя
	userUUID = uuid.New()

	// 4. Сохраняем пользователя
	if err := svc.uw.Save(ctx, userUUID, username, passwordHash); err != nil {
		return uuid.Nil, err
	}

	// 5. Возвращаем UUID пользователя
	return userUUID, nil
}

// Login проверяет пользователя и устройство, а затем возвращает JWT-токен.
func (svc *AuthService) Login(
	ctx context.Context,
	username string,
	password string,
	deviceUUID uuid.UUID,
) (token string, err error) {
	// 1. Получаем пользователя
	user, err := svc.ur.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.ErrUserNotFound
	}

	// 2. Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.ErrInvalidPassword
	}

	// 3. Проверяем устройство
	device, err := svc.dr.GetByUUID(ctx, deviceUUID)
	if err != nil {
		return "", err
	}
	if device == nil || device.UserUUID != user.UserUUID {
		return "", errors.ErrDeviceNotFound
	}

	// 4. Генерируем токен
	token, err = svc.tg.Generate(user.UserUUID, deviceUUID)
	if err != nil {
		return "", err
	}

	return token, nil
}

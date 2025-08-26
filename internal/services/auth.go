package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Ошибки сервиса
var (
	ErrUserExists        = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user does not exist")
	ErrDeviceNotFound    = errors.New("device does not exist")
	ErrInvalidCredential = errors.New("invalid credentials")
)

// UserWriter интерфейс для записи пользователей в БД
type UserWriter interface {
	Save(ctx context.Context, userUUID uuid.UUID, username, password string) error
}

// UserReader интерфейс для чтения пользователей из БД
type UserReader interface {
	GetByUsername(ctx context.Context, username string) (*models.UserDB, error)
}

// DeviceWriter интерфейс для записи устройств в БД
type DeviceWriter interface {
	Save(ctx context.Context, deviceUUID, userUUID uuid.UUID, publicKey string) error
}

// DeviceReader интерфейс для чтения устройств из БД
type DeviceReader interface {
	Get(ctx context.Context, deviceUUID uuid.UUID) (*models.DeviceDB, error)
}

// TokenGenerator интерфейс для генерации JWT токенов
type TokenGenerator interface {
	Generate(userUUID uuid.UUID, deviceUUID uuid.UUID) (string, error)
}

// AuthService сервис авторизации
type AuthService struct {
	userWriteRepo   UserWriter
	userReadRepo    UserReader
	deviceWriteRepo DeviceWriter
	deviceReadRepo  DeviceReader
	tokenGen        TokenGenerator
}

// NewAuthService создаёт новый сервис авторизации
func NewAuthService(
	userWriteRepo UserWriter,
	userReadRepo UserReader,
	deviceWriteRepo DeviceWriter,
	deviceReadRepo DeviceReader,
	tokenGen TokenGenerator,
) *AuthService {
	return &AuthService{
		userWriteRepo:   userWriteRepo,
		userReadRepo:    userReadRepo,
		deviceWriteRepo: deviceWriteRepo,
		deviceReadRepo:  deviceReadRepo,
		tokenGen:        tokenGen,
	}
}

// Register создаёт нового пользователя
func (s *AuthService) Register(ctx context.Context, username, password string) (userUUID uuid.UUID, err error) {
	var user *models.UserDB
	user, err = s.userReadRepo.GetByUsername(ctx, username)
	if err != nil {
		return
	}
	if user != nil {
		err = ErrUserExists
		return
	}

	var hash []byte
	hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return
	}

	userUUID = uuid.New()
	err = s.userWriteRepo.Save(ctx, userUUID, username, string(hash))
	return
}

// AddDevice добавляет новое устройство для пользователя по username/password
func (s *AuthService) AddDevice(ctx context.Context, username, password, publicKey string) (deviceUUID uuid.UUID, err error) {
	var user *models.UserDB
	user, err = s.userReadRepo.GetByUsername(ctx, username)
	if err != nil {
		return
	}
	if user == nil {
		err = ErrUserNotFound
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		err = ErrInvalidCredential
		return
	}

	deviceUUID = uuid.New()
	err = s.deviceWriteRepo.Save(ctx, deviceUUID, user.UserUUID, publicKey)
	if err != nil {
		deviceUUID = uuid.Nil
		return
	}
	return
}

// Login проверяет логин и пароль, возвращает JWT токен
func (s *AuthService) Login(ctx context.Context, username, password string, deviceUUID uuid.UUID) (tokenString string, err error) {
	var user *models.UserDB
	user, err = s.userReadRepo.GetByUsername(ctx, username)
	if err != nil {
		return
	}
	if user == nil {
		err = ErrUserNotFound
		return
	}

	var device *models.DeviceDB
	device, err = s.deviceReadRepo.Get(ctx, deviceUUID)
	if err != nil {
		return
	}
	if device == nil {
		err = ErrDeviceNotFound
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		err = ErrInvalidCredential
		return
	}

	tokenString, err = s.tokenGen.Generate(user.UserUUID, deviceUUID)
	return
}

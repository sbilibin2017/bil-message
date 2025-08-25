package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// Ошибки
var (
	// ErrUsernameAlreadyExists возвращается при попытке регистрации уже существующего пользователя
	ErrUsernameAlreadyExists = errors.New("username already exists")

	// ErrInvalidCredentials возвращается, если переданы неверные имя пользователя или пароль
	ErrInvalidCredentials = errors.New("invalid username or password")
)

//
// Интерфейсы
//

// UserGetter описывает интерфейс получения пользователя по username
type UserGetter interface {
	// Get возвращает пользователя по имени или nil, если такого пользователя нет
	Get(ctx context.Context, username string) (*models.UserDB, error)
}

// UserSaver описывает интерфейс сохранения нового пользователя
type UserSaver interface {
	// Save сохраняет пользователя с UUID, username и хэшем пароля
	Save(ctx context.Context, userUUID uuid.UUID, username string, passwordHash string) error
}

// DeviceGetter
type DeviceGetter interface {
	Get(ctx context.Context, deviceUUID uuid.UUID) (*models.UserDeviceDB, error)
}

// DeviceSaver описывает интерфейс сохранения нового устройства пользователя
type DeviceSaver interface {
	// Save сохраняет устройство с UUID, привязанное к пользователю и его публичному ключу
	Save(ctx context.Context, deviceUUID uuid.UUID, userUUID uuid.UUID, publicKey string) error
}

// TokenGenerator описывает интерфейс генерации JWT токена
type TokenGenerator interface {
	// Generate создает JWT токен, содержащий userUUID и deviceUUID
	Generate(userUUID uuid.UUID, deviceUUID uuid.UUID) (string, error)
}

//
// Сервис аутентификации и управления пользователями/устройствами
//

// AuthService предоставляет методы для:
//   - регистрации пользователей
//   - добавления устройств
//   - входа (логина) с проверкой пароля и генерацией токена
type AuthService struct {
	ug UserGetter
	us UserSaver
	dg DeviceGetter
	ds DeviceSaver
	tg TokenGenerator
}

// NewAuthService создаёт новый экземпляр AuthService
func NewAuthService(
	ug UserGetter,
	us UserSaver,
	dg DeviceGetter,
	ds DeviceSaver,
	tg TokenGenerator,
) *AuthService {
	return &AuthService{
		ug: ug,
		us: us,
		dg: dg,
		ds: ds,
		tg: tg,
	}
}

// Register создаёт нового пользователя с указанными username и password.
// Если пользователь уже существует, возвращает ErrUsernameAlreadyExists.
// Пароль хэшируется с использованием bcrypt.
func (svc *AuthService) Register(
	ctx context.Context,
	username string,
	password string,
) (userUUID uuid.UUID, err error) {
	// Проверяем, существует ли пользователь
	existing, err := svc.ug.Get(ctx, username)
	if err != nil {
		return uuid.Nil, err
	}
	if existing != nil {
		return uuid.Nil, ErrUsernameAlreadyExists
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.Nil, err
	}

	userUUID = uuid.New()

	// Сохраняем пользователя
	err = svc.us.Save(ctx, userUUID, username, string(hashedPassword))
	if err != nil {
		return uuid.Nil, err
	}
	return userUUID, nil
}

// AddDevice добавляет новое устройство пользователю.
// Проверяет логин/пароль, создает UUID для устройства, сохраняет его в БД вместе с publicKey.
// Возвращает UUID устройства. JWT не создается (это делает Login).
func (svc *AuthService) AddDevice(
	ctx context.Context,
	username string,
	password string,
	publicKey string,
) (deviceUUID uuid.UUID, err error) {
	// Проверяем пользователя
	user, err := svc.ug.Get(ctx, username)
	if err != nil {
		return uuid.Nil, err
	}
	if user == nil {
		return uuid.Nil, ErrInvalidCredentials
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return uuid.Nil, ErrInvalidCredentials
	}

	// Генерируем UUID устройства
	deviceUUID = uuid.New()

	// Сохраняем устройство в базе
	if err := svc.ds.Save(ctx, deviceUUID, user.UserUUID, publicKey); err != nil {
		return uuid.Nil, err
	}

	return deviceUUID, nil
}

// Login проверяет учетные данные пользователя и выдает JWT для конкретного устройства.
// Для входа клиент передает username, password и deviceUUID.
// Если пароль неверный или пользователя нет, возвращает ErrInvalidCredentials.
// Login проверяет учетные данные пользователя и выдает JWT для конкретного устройства.
// Проверяет, что устройство существует и принадлежит пользователю.
// Если пароль неверный, устройство не найдено или пользователь не существует, возвращает ErrInvalidCredentials.
func (svc *AuthService) Login(
	ctx context.Context,
	username string,
	password string,
	deviceUUID uuid.UUID,
) (token string, err error) {
	// Находим пользователя
	user, err := svc.ug.Get(ctx, username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// Проверяем устройство
	device, err := svc.dg.Get(ctx, deviceUUID)
	if err != nil {
		return "", err
	}
	if device == nil || device.UserUUID != user.UserUUID {
		return "", ErrInvalidCredentials
	}

	// Генерируем JWT для указанного устройства
	token, err = svc.tg.Generate(user.UserUUID, deviceUUID)
	if err != nil {
		return "", err
	}

	return token, nil
}

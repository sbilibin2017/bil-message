package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// ErrUserAlreadyExists возвращается, если пользователь с таким username уже существует.
var ErrUserAlreadyExists = errors.New("user already exists")

// UserReader определяет интерфейс для чтения данных пользователя из хранилища.
type UserReader interface {
	// Get возвращает пользователя по username.
	Get(ctx context.Context, username string) (*models.UserDB, error)
}

// UserWriter определяет интерфейс для сохранения данных пользователя в хранилище.
type UserWriter interface {
	// Save сохраняет пользователя с заданным UUID, именем и хэшем пароля.
	Save(ctx context.Context, userUUID uuid.UUID, username string, passwordHash string) error
}

// TokenGenerator определяет интерфейс для генерации токенов.
type TokenGenerator interface {
	// Generate создает токен для указанного клиента и пользователя.
	Generate(clientUUID uuid.UUID, userUUID uuid.UUID) (string, error)
}

// AuthService предоставляет методы для создания пользователей и клиентов.
type AuthService struct {
	ur UserReader
	uw UserWriter
	tg TokenGenerator
}

// NewAuthService создаёт новый экземпляр AuthService.
func NewAuthService(
	ur UserReader,
	uw UserWriter,
	tg TokenGenerator,
) *AuthService {
	return &AuthService{
		ur: ur,
		uw: uw,
		tg: tg,
	}
}

// Register создаёт нового пользователя, клиента с RSA ключами и возвращает токен и приватный ключ.
func (svc *AuthService) Register(
	ctx context.Context,
	username string,
	password string,
) (token string, err error) {
	// 1. Проверяем, существует ли пользователь
	existingUser, err := svc.ur.Get(ctx, username)
	if err != nil {
		return "", err
	}
	if existingUser != nil {
		return "", ErrUserAlreadyExists
	}

	// 2. Хэшируем пароль
	passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	passwordHash := string(passwordHashBytes)

	// 3. Генерируем UUID для пользователя и клиента
	userUUID := uuid.New()
	clientUUID := uuid.New()

	// 5. Сохраняем пользователя
	if err := svc.uw.Save(ctx, userUUID, username, passwordHash); err != nil {
		return "", err
	}

	// 7. Генерируем JWT токен
	token, err = svc.tg.Generate(clientUUID, userUUID)
	if err != nil {
		return "", err
	}

	// 8. Возвращаем токен и приватный ключ пользователю
	return token, nil
}

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
	Get(
		ctx context.Context,
		username string,
	) (*models.UserDB, error)
}

// UserWriter определяет интерфейс для сохранения данных пользователя в хранилище.
type UserWriter interface {
	Save(
		ctx context.Context,
		userUUID string,
		username string,
		password_hash string,
	) error
}

// TokenGenerator
type TokenGenerator interface {
	Generate(userUUID string) (string, error)
}

// AuthService предоставляет методы для создания пользователей и клиентов.
type AuthService struct {
	ur UserReader
	uw UserWriter
	tg TokenGenerator
}

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

// Register создаёт нового пользователя и возвращает его UUID.
func (svc *AuthService) Register(ctx context.Context, username, password string) (tokenString string, err error) {
	// Проверяем, существует ли пользователь
	existingUser, err := svc.ur.Get(ctx, username)
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

	// Генерируем UUID пользователя
	userUUID := uuid.New().String()

	// Сохраняем пользователя через интерфейс UserWriter
	if err := svc.uw.Save(ctx, userUUID, username, string(hashBytes)); err != nil {
		return "", err
	}

	// Генерируем токен через интерфейс TokenGenerator
	token, err := svc.tg.Generate(userUUID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Login проверяет пользователя и возвращает токен.
func (svc *AuthService) Login(ctx context.Context, username, password string) (tokenString string, err error) {
	// Получаем пользователя
	user, err := svc.ur.Get(ctx, username)
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

	// Генерируем токен через интерфейс TokenGenerator
	token, err := svc.tg.Generate(user.UserUUID)
	if err != nil {
		return "", err
	}

	return token, nil
}

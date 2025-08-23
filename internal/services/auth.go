package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// ErrUsernameAlreadyExists возвращается, если username уже существует
var ErrUsernameAlreadyExists = errors.New("username already exists")

// UserGetter описывает интерфейс получения пользователя по username
type UserGetter interface {
	// Get возвращает пользователя по username или nil, если пользователь не найден
	Get(ctx context.Context, username string) (*models.UserDB, error)
}

// UserSaver описывает интерфейс сохранения нового пользователя
type UserSaver interface {
	// Save сохраняет пользователя в базу данных
	Save(ctx context.Context, userUUID uuid.UUID, username string, passwordHash string) error
}

// AuthService предоставляет методы аутентификации и регистрации пользователей
type AuthService struct {
	ug UserGetter
	us UserSaver
}

// NewAuthService создаёт новый экземпляр AuthService
func NewAuthService(
	ug UserGetter,
	us UserSaver,
) *AuthService {
	return &AuthService{ug: ug, us: us}
}

// Register создаёт нового пользователя с указанными username и password
// Возвращает сгенерированный UUID пользователя или ошибку
func (svc *AuthService) Register(
	ctx context.Context,
	username string,
	password string,
) (userUUID uuid.UUID, err error) {
	// Проверяем, существует ли пользователь с таким username
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

	// Генерируем UUID нового пользователя
	userUUID = uuid.New()

	// Сохраняем пользователя в базе
	if err := svc.us.Save(ctx, userUUID, username, string(hashedPassword)); err != nil {
		return uuid.Nil, err
	}

	return userUUID, nil
}

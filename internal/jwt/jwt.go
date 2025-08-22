package jwt

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userCtxKey contextKey = "user"

// JWT представляет работу с JWT-токенами.
type JWT struct {
	secretKey []byte
	exp       time.Duration
}

// Opt — функциональная опция для настройки JWT.
type Opt func(*JWT) error

// New создаёт новый JWT, применяя указанные опции.
func New(opts ...Opt) (*JWT, error) {
	j := &JWT{
		secretKey: []byte("secret-key"),
		exp:       time.Hour,
	}
	for _, opt := range opts {
		if err := opt(j); err != nil {
			return nil, err
		}
	}
	return j, nil
}

// claims — приватная структура для хранения payload токена.
type claims struct {
	UserUUID string `json:"user_uuid"`
	jwt.RegisteredClaims
}

// WithSecretKey задаёт секретный ключ.
func WithSecretKey(secret ...string) Opt {
	return func(j *JWT) error {
		for _, s := range secret {
			if s != "" {
				j.secretKey = []byte(s)
				return nil
			}
		}
		return nil
	}
}

// WithExpiration задаёт время жизни токена.
func WithExpiration(exp ...time.Duration) Opt {
	return func(j *JWT) error {
		for _, e := range exp {
			if e > 0 {
				j.exp = e
				return nil
			}
		}
		return nil
	}
}

// Generate создаёт JWT-токен на основе userUUID.
func (j *JWT) Generate(userUUID string) (string, error) {
	c := claims{
		UserUUID: userUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(j.secretKey)
}

// GetFromRequest извлекает JWT-токен из заголовка Authorization.
func (j *JWT) GetFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing Authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid Authorization header format")
	}

	return parts[1], nil
}

// GetUserUUID проверяет токен и возвращает userUUID.
func (j *JWT) GetUserUUID(tokenString string) (string, error) {
	c := &claims{}

	token, err := jwt.ParseWithClaims(tokenString, c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}

	return c.UserUUID, nil
}

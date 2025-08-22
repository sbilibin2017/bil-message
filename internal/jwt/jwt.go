package jwt

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sbilibin2017/bil-message/internal/models"
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

// WithSecretKey задаёт секретный ключ.
// Используется первое непустое значение.
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
// Используется первое положительное значение.
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

// Generate создаёт JWT-токен на основе TokenPayload.
func (j *JWT) Generate(payload *models.TokenPayload) (string, error) {
	c := models.Claims{
		UserUUID:   payload.UserUUID,
		DeviceUUID: payload.DeviceUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString(j.secretKey)
}

// GetFromRequest извлекает JWT токен из заголовка Authorization запроса.
func (j *JWT) GetFromRequest(r *http.Request) (*string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("missing Authorization header")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("invalid Authorization header format")
	}

	return &parts[1], nil
}

// Parse проверяет токен и возвращает TokenPayload.
func (j *JWT) Parse(tokenString string) (*models.TokenPayload, error) {
	claims := &models.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &models.TokenPayload{
		UserUUID:   claims.UserUUID,
		DeviceUUID: claims.DeviceUUID,
	}, nil
}

// SetToContext сохраняет TokenPayload в контекст.
func (j *JWT) SetToContext(ctx context.Context, payload *models.TokenPayload) context.Context {
	return context.WithValue(ctx, userCtxKey, payload)
}

// GetTokenPayloadFromContext извлекает TokenPayload из контекста.
func (j *JWT) GetTokenPayloadFromContext(ctx context.Context) (*models.TokenPayload, error) {
	value := ctx.Value(userCtxKey)
	if value == nil {
		return nil, errors.New("no token payload in context")
	}

	payload, ok := value.(*models.TokenPayload)
	if !ok {
		return nil, errors.New("invalid token payload type")
	}

	return payload, nil
}

package jwt

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type contextKey string

const userCtxKey contextKey = "user"

// JWT представляет работу с JWT-токенами.
type JWT struct {
	secretKey []byte
	exp       time.Duration
}

// New создаёт новый JWT с указанным секретным ключом и временем жизни токена.
func New(secretKey string, exp time.Duration) (*JWT, error) {
	if secretKey == "" {
		return nil, errors.New("secret key cannot be empty")
	}
	if exp <= 0 {
		exp = time.Hour
	}
	return &JWT{
		secretKey: []byte(secretKey),
		exp:       exp,
	}, nil
}

type claimsStruct struct {
	UserUUID   string `json:"user_uuid"`
	ClientUUID string `json:"client_uuid"`
	jwt.RegisteredClaims
}

// Generate создаёт JWT токен на основе userUUID и clientUUID.
func (j *JWT) Generate(userUUID uuid.UUID, clientUUID uuid.UUID) (string, error) {
	claims := claimsStruct{
		UserUUID:   userUUID.String(),
		ClientUUID: clientUUID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.exp)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// GetFromRequest извлекает JWT токен из заголовка Authorization запроса.
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

// Parse проверяет токен и возвращает userUUID и clientUUID.
func (j *JWT) Parse(tokenString string) (userUUID uuid.UUID, clientUUID uuid.UUID, err error) {
	claims := &claimsStruct{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, uuid.Nil, errors.New("invalid token")
	}

	userUUID, err = uuid.Parse(claims.UserUUID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	clientUUID, err = uuid.Parse(claims.ClientUUID)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	return userUUID, clientUUID, nil
}

// SetToContext сохраняет userUUID и clientUUID напрямую в контекст.
func (j *JWT) SetToContext(ctx context.Context, userUUID uuid.UUID, clientUUID uuid.UUID) context.Context {
	return context.WithValue(ctx, userCtxKey, [2]uuid.UUID{userUUID, clientUUID})
}

// GetTokenPayloadFromContext извлекает userUUID и clientUUID напрямую из контекста.
func (j *JWT) GetTokenPayloadFromContext(ctx context.Context) (userUUID uuid.UUID, clientUUID uuid.UUID, err error) {
	value := ctx.Value(userCtxKey)
	if value == nil {
		return uuid.Nil, uuid.Nil, errors.New("no token payload in context")
	}

	uuids, ok := value.([2]uuid.UUID)
	if !ok {
		return uuid.Nil, uuid.Nil, errors.New("invalid token payload type")
	}

	return uuids[0], uuids[1], nil
}

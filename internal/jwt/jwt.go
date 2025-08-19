package jwt

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// contextKey — собственный тип для ключей в контексте, чтобы избежать коллизий.
type contextKey string

const userCtxKey contextKey = "user"

// JWT представляет работу с JWT-токенами.
type JWT struct {
	SecretKey []byte
}

// New создаёт новый экземпляр JWT с указанным секретным ключом.
func New(secretKey string) *JWT {
	return &JWT{
		SecretKey: []byte(secretKey),
	}
}

// claimsStruct — приватная структура для хранения claims JWT.
type claimsStruct struct {
	UserUUID   string `json:"user_uuid"`
	ClientUUID string `json:"client_uuid"`
	jwt.RegisteredClaims
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

// Parse проверяет токен и возвращает его payload в виде TokenPayload.
func (j *JWT) Parse(tokenString string) (*models.TokenPayload, error) {
	claims := &claimsStruct{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.SecretKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	userUUID, err := uuid.Parse(claims.UserUUID)
	if err != nil {
		return nil, err
	}

	clientUUID, err := uuid.Parse(claims.ClientUUID)
	if err != nil {
		return nil, err
	}

	return &models.TokenPayload{
		UserUUID:   userUUID,
		ClientUUID: clientUUID,
	}, nil
}

// SetToContext сохраняет payload в контекст запроса для дальнейшего использования.
func (j *JWT) SetToContext(ctx context.Context, payload *models.TokenPayload) (context.Context, error) {
	if payload == nil {
		return ctx, errors.New("payload is nil")
	}
	return context.WithValue(ctx, userCtxKey, payload), nil
}

// GetTokenPayloadFromContext извлекает TokenPayload из контекста.
func (j *JWT) GetTokenPayloadFromContext(ctx context.Context) (*models.TokenPayload, error) {
	payload, ok := ctx.Value(userCtxKey).(*models.TokenPayload)
	if !ok || payload == nil {
		return nil, errors.New("no token payload in context")
	}
	return payload, nil
}

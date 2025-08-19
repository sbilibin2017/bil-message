package jwt

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/assert"
)

// --------------------
// Test GetFromRequest
// --------------------
func TestJWT_GetFromRequest_Success(t *testing.T) {
	j := New("secret")
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer mytoken123")

	token, err := j.GetFromRequest(req)
	assert.NoError(t, err)
	assert.Equal(t, "mytoken123", token)
}

func TestJWT_GetFromRequest_MissingHeader(t *testing.T) {
	j := New("secret")
	req, _ := http.NewRequest("GET", "/", nil)

	token, err := j.GetFromRequest(req)
	assert.Error(t, err)
	assert.Equal(t, "", token)
}

func TestJWT_GetFromRequest_InvalidFormat(t *testing.T) {
	j := New("secret")
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Token mytoken123")

	token, err := j.GetFromRequest(req)
	assert.Error(t, err)
	assert.Equal(t, "", token)
}

// --------------------
// Test Parse
// --------------------
func TestJWT_Parse_ValidToken(t *testing.T) {
	secret := "secret"
	j := New(secret)

	userUUID := uuid.New()
	clientUUID := uuid.New()

	claims := jwt.MapClaims{
		"user_uuid":   userUUID.String(),
		"client_uuid": clientUUID.String(),
		"exp":         time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	payload, err := j.Parse(tokenString)
	assert.NoError(t, err)
	assert.Equal(t, userUUID, payload.UserUUID)
	assert.Equal(t, clientUUID, payload.ClientUUID)
}

func TestJWT_Parse_InvalidToken(t *testing.T) {
	j := New("secret")
	_, err := j.Parse("invalid.token")
	assert.Error(t, err)
}

func TestJWT_Parse_WrongSecret(t *testing.T) {
	secret := "secret"
	j := New("wrongsecret")

	userUUID := uuid.New()
	clientUUID := uuid.New()
	claims := jwt.MapClaims{
		"user_uuid":   userUUID.String(),
		"client_uuid": clientUUID.String(),
		"exp":         time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	_, err := j.Parse(tokenString)
	assert.Error(t, err)
}

// --------------------
// Test SetToContext
// --------------------
func TestJWT_SetToContext_Success(t *testing.T) {
	j := New("secret")
	ctx := context.Background()
	payload := &models.TokenPayload{
		UserUUID:   uuid.New(),
		ClientUUID: uuid.New(),
	}

	newCtx, err := j.SetToContext(ctx, payload)
	assert.NoError(t, err)
	assert.NotNil(t, newCtx)
}

func TestJWT_SetToContext_NilPayload(t *testing.T) {
	j := New("secret")
	ctx := context.Background()
	newCtx, err := j.SetToContext(ctx, nil)
	assert.Error(t, err)
	assert.Equal(t, ctx, newCtx)
}

// --------------------
// Test GetTokenPayloadFromContext
// --------------------
func TestJWT_GetTokenPayloadFromContext_Success(t *testing.T) {
	j := New("secret")
	payload := &models.TokenPayload{
		UserUUID:   uuid.New(),
		ClientUUID: uuid.New(),
	}
	ctx, _ := j.SetToContext(context.Background(), payload)

	gotPayload, err := j.GetTokenPayloadFromContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, payload, gotPayload)
}

func TestJWT_GetTokenPayloadFromContext_EmptyContext(t *testing.T) {
	j := New("secret")
	_, err := j.GetTokenPayloadFromContext(context.Background())
	assert.Error(t, err)
}

func TestJWT_GetTokenPayloadFromContext_WrongType(t *testing.T) {
	j := New("secret")
	ctx := context.WithValue(context.Background(), userCtxKey, "not_a_payload")
	_, err := j.GetTokenPayloadFromContext(ctx)
	assert.Error(t, err)
}

package jwt

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewJWT(t *testing.T) {
	j, err := New("secret", time.Minute)
	assert.NoError(t, err)
	assert.NotNil(t, j)

	_, err = New("", time.Minute)
	assert.Error(t, err)

	j2, err := New("secret", 0) // проверка default exp
	assert.NoError(t, err)
	assert.NotNil(t, j2)
}

func TestGenerateAndParse(t *testing.T) {
	j, _ := New("secret", time.Minute)
	userUUID := uuid.New()
	clientUUID := uuid.New()

	token, err := j.Generate(userUUID, clientUUID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedUser, parsedClient, err := j.Parse(token)
	assert.NoError(t, err)
	assert.Equal(t, userUUID, parsedUser)
	assert.Equal(t, clientUUID, parsedClient)
}

func TestGetFromRequest(t *testing.T) {
	j, _ := New("secret", time.Minute)

	req := &http.Request{Header: http.Header{}}

	_, err := j.GetFromRequest(req)
	assert.Error(t, err) // нет заголовка

	req.Header.Set("Authorization", "Bearer token123")
	token, err := j.GetFromRequest(req)
	assert.NoError(t, err)
	assert.Equal(t, "token123", token)

	req.Header.Set("Authorization", "InvalidFormat")
	_, err = j.GetFromRequest(req)
	assert.Error(t, err)
}

func TestContextMethods(t *testing.T) {
	j, _ := New("secret", time.Minute)
	ctx := context.Background()

	userUUID := uuid.New()
	clientUUID := uuid.New()

	ctxWithValue := j.SetToContext(ctx, userUUID, clientUUID)
	gotUser, gotClient, err := j.GetTokenPayloadFromContext(ctxWithValue)
	assert.NoError(t, err)
	assert.Equal(t, userUUID, gotUser)
	assert.Equal(t, clientUUID, gotClient)

	// Тест ошибки при пустом контексте
	_, _, err = j.GetTokenPayloadFromContext(ctx)
	assert.Error(t, err)
}

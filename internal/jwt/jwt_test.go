package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
)

func TestJWT_GenerateAndParse(t *testing.T) {
	j, err := New()
	assert.NoError(t, err)

	userID := uuid.New()
	deviceID := uuid.New()

	token, err := j.Generate(userID, deviceID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsedUserID, parsedDeviceID, err := j.Parse(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, parsedUserID)
	assert.Equal(t, deviceID, parsedDeviceID)
}

func TestJWT_Parse_InvalidToken(t *testing.T) {
	j, _ := New()

	_, _, err := j.Parse("invalid.token.value")
	assert.Error(t, err)
}

func TestJWT_WithSecretKeyAndExpiration(t *testing.T) {
	j, err := New(
		WithSecretKey("mysecret"),
		WithExpiration(2*time.Hour),
	)
	assert.NoError(t, err)
	assert.NotNil(t, j)

	userID := uuid.New()
	deviceID := uuid.New()

	token, err := j.Generate(userID, deviceID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

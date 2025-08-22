package jwt

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/require"
)

func TestNew_Defaults(t *testing.T) {
	j, err := New()
	require.NoError(t, err)
	require.NotNil(t, j)
	require.Equal(t, time.Hour, j.exp)
	require.Equal(t, []byte("secret-key"), j.secretKey)
}

func TestNew_WithSecretKey(t *testing.T) {
	j, err := New(WithSecretKey("", "custom-key"))
	require.NoError(t, err)
	require.Equal(t, []byte("custom-key"), j.secretKey)
}

func TestNew_WithExpiration(t *testing.T) {
	j, err := New(WithExpiration(0, 2*time.Hour))
	require.NoError(t, err)
	require.Equal(t, 2*time.Hour, j.exp)
}

func TestGenerateAndParse(t *testing.T) {
	j, err := New(WithSecretKey("test-secret"), WithExpiration(time.Minute))
	require.NoError(t, err)

	payload := &models.TokenPayload{
		UserUUID:   uuid.NewString(),
		DeviceUUID: uuid.NewString(),
	}

	token, err := j.Generate(payload)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	parsed, err := j.Parse(token)
	require.NoError(t, err)
	require.Equal(t, payload.UserUUID, parsed.UserUUID)
	require.Equal(t, payload.DeviceUUID, parsed.DeviceUUID)
}

func TestParse_InvalidToken(t *testing.T) {
	j, _ := New(WithSecretKey("test-secret"))
	parsed, err := j.Parse("invalid.token.string")
	require.Error(t, err)
	require.Nil(t, parsed)
}

func TestGetFromRequest(t *testing.T) {
	j, _ := New()

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	token, err := j.GetFromRequest(req)
	require.NoError(t, err)
	require.NotNil(t, token)
	require.Equal(t, "test-token", *token)
}

func TestGetFromRequest_MissingHeader(t *testing.T) {
	j, _ := New()

	req, _ := http.NewRequest("GET", "/", nil)
	token, err := j.GetFromRequest(req)
	require.Error(t, err)
	require.Nil(t, token)
}

func TestContext(t *testing.T) {
	j, _ := New()
	payload := &models.TokenPayload{
		UserUUID:   uuid.NewString(),
		DeviceUUID: uuid.NewString(),
	}

	ctx := j.SetToContext(context.Background(), payload)
	got, err := j.GetTokenPayloadFromContext(ctx)
	require.NoError(t, err)
	require.Equal(t, payload, got)
}

func TestContext_NoPayload(t *testing.T) {
	j, _ := New()
	_, err := j.GetTokenPayloadFromContext(context.Background())
	require.Error(t, err)
}

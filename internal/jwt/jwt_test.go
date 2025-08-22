package jwt_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/sbilibin2017/bil-message/internal/jwt"
	"github.com/stretchr/testify/assert"
)

func TestWithSecretKey(t *testing.T) {
	// Создаём новый JWT с дефолтным ключом
	_, err := jwt.New()
	assert.NoError(t, err)

	// Создаём новый JWT с опцией WithSecretKey
	secret := "my-secret-key"
	_, err = jwt.New(jwt.WithSecretKey(secret))
	assert.NoError(t, err)

	// Если передать несколько значений, берётся первое непустое
	_, err = jwt.New(jwt.WithSecretKey("", "", "first-non-empty", "ignored"))
	assert.NoError(t, err)

	// Если все значения пустые, ключ остаётся по умолчанию
	_, err = jwt.New(jwt.WithSecretKey("", ""))
	assert.NoError(t, err)
}

func TestJWT_GenerateAndParse(t *testing.T) {
	j, err := jwt.New()
	assert.NoError(t, err)

	userUUID := "user-123"

	// Генерация токена
	tokenStr, err := j.Generate(userUUID)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenStr)

	// Разбор токена
	parsedUUID, err := j.GetUserUUID(tokenStr)
	assert.NoError(t, err)
	assert.Equal(t, userUUID, parsedUUID)
}

func TestJWT_GetFromRequest(t *testing.T) {
	j, _ := jwt.New()
	userUUID := "user-123"
	tokenStr, _ := j.Generate(userUUID)

	req := &http.Request{
		Header: map[string][]string{
			"Authorization": {"Bearer " + tokenStr},
		},
	}

	got, err := j.GetFromRequest(req)
	assert.NoError(t, err)
	assert.Equal(t, tokenStr, got)
}

func TestJWT_GetUserUUID_InvalidToken(t *testing.T) {
	j, _ := jwt.New()
	invalidToken := "invalid.token.value"

	got, err := j.GetUserUUID(invalidToken)
	assert.Error(t, err)
	assert.Empty(t, got)
}

func TestJWT_ExpiredToken(t *testing.T) {
	j, _ := jwt.New(jwt.WithExpiration(1 * time.Millisecond))
	userUUID := "user-expired"

	tokenStr, err := j.Generate(userUUID)
	assert.NoError(t, err)

	time.Sleep(2 * time.Millisecond) // ждем истечения срока

	got, err := j.GetUserUUID(tokenStr)
	assert.Error(t, err)
	assert.Empty(t, got)
}

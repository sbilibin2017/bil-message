package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	// Создаем тестовый HTTP сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем URL и метод
		if r.URL.Path != "/auth/register" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Читаем тело запроса (не обязательно для теста, но можно проверить)
		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Отвечаем фиктивным UUID
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "123e4567-e89b-12d3-a456-426614174000")
	}))
	defer ts.Close()

	// Создаем resty клиент для тестового сервера
	restClient := resty.New().SetBaseURL(ts.URL)

	// Вызываем функцию Register
	username := "testuser"
	password := "testpass"
	uuid, err := Register(context.Background(), restClient, username, password)

	// Проверяем ошибки и результат
	assert.NoError(t, err)
	assert.NotNil(t, uuid)
	assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", *uuid)
}

func TestRegister_NetworkError(t *testing.T) {
	// Используем недоступный URL, чтобы вызвать ошибку сети
	restClient := resty.New().SetBaseURL("http://127.0.0.1:0") // порт 0 не существует

	username := "user"
	password := "pass"

	uuid, err := Register(context.Background(), restClient, username, password)

	assert.Nil(t, uuid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestRegister_ServerError(t *testing.T) {
	// Создаем тестовый сервер, который всегда возвращает 500
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	restClient := resty.New().SetBaseURL(ts.URL)
	username := "user"
	password := "pass"

	uuid, err := Register(context.Background(), restClient, username, password)

	assert.Nil(t, uuid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server returned error")
}

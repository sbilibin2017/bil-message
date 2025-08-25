package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	// Создаем тестовый HTTP сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/register" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "123e4567-e89b-12d3-a456-426614174000") // UUID is ignored
	}))
	defer ts.Close()

	restClient := resty.New().SetBaseURL(ts.URL)

	err := Register(context.Background(), restClient, "testuser", "testpass")

	assert.NoError(t, err)
}

func TestRegister_NetworkError(t *testing.T) {
	restClient := resty.New().SetBaseURL("http://127.0.0.1:0") // недоступный адрес

	err := Register(context.Background(), restClient, "user", "pass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestRegister_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	restClient := resty.New().SetBaseURL(ts.URL)

	err := Register(context.Background(), restClient, "user", "pass")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "server returned error")
}

func TestAddDevice(t *testing.T) {
	// Создаем тестовый сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/device" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Возвращаем фиктивный UUID устройства
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "123e4567-e89b-12d3-a456-426614174001")
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	deviceUUID, err := AddDevice(context.Background(), client, "user", "pass", "pubkey")

	assert.NoError(t, err)
	assert.Equal(t, uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), deviceUUID)
}

func TestAddDevice_NetworkError(t *testing.T) {
	client := resty.New().SetBaseURL("http://127.0.0.1:0") // недоступный адрес

	deviceUUID, err := AddDevice(context.Background(), client, "user", "pass", "pubkey")

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, deviceUUID)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestAddDevice_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	deviceUUID, err := AddDevice(context.Background(), client, "user", "pass", "pubkey")

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, deviceUUID)
	assert.Contains(t, err.Error(), "server returned error")
}

func TestLogin(t *testing.T) {
	// Тестовый сервер возвращает JWT в заголовке Authorization
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/login" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Authorization", "Bearer faketoken123")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	token, err := Login(context.Background(), client, "user", "pass", uuid.New())

	assert.NoError(t, err)
	assert.Equal(t, "Bearer faketoken123", token)
}

func TestLogin_NetworkError(t *testing.T) {
	client := resty.New().SetBaseURL("http://127.0.0.1:0")

	token, err := Login(context.Background(), client, "user", "pass", uuid.New())

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestLogin_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	token, err := Login(context.Background(), client, "user", "pass", uuid.New())

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "server returned error")
}

func TestLogin_NoAuthHeader(t *testing.T) {
	// Сервер возвращает 200, но без заголовка Authorization
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	token, err := Login(context.Background(), client, "user", "pass", uuid.New())

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "no Authorization header returned")
}

func TestAddDevice_InvalidUUID(t *testing.T) {
	// Создаем тестовый сервер, который возвращает некорректный UUID
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "not-a-valid-uuid")
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	deviceUUID, err := AddDevice(context.Background(), client, "user", "pass", "pubkey")

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, deviceUUID)
	assert.Contains(t, err.Error(), "invalid device UUID returned")
}

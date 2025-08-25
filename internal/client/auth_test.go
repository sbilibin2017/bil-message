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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/register" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "123e4567-e89b-12d3-a456-426614174000")
	}))
	defer ts.Close()

	restClient := resty.New().SetBaseURL(ts.URL)

	userUUID, err := Register(context.Background(), restClient, "testuser", "testpass")

	assert.NoError(t, err)
	assert.Equal(t, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), userUUID)
}

func TestRegister_NetworkError(t *testing.T) {
	restClient := resty.New().SetBaseURL("http://127.0.0.1:0") // недоступный адрес

	userUUID, err := Register(context.Background(), restClient, "user", "pass")

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, userUUID)
	assert.Contains(t, err.Error(), "failed to send request")
}

func TestRegister_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	restClient := resty.New().SetBaseURL(ts.URL)

	userUUID, err := Register(context.Background(), restClient, "user", "pass")

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, userUUID)
	assert.Contains(t, err.Error(), "server returned error")
}

func TestAddDevice(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/device" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "123e4567-e89b-12d3-a456-426614174001")
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	deviceUUID, err := AddDevice(context.Background(), client, "user", "pass", "pubkey")

	assert.NoError(t, err)
	assert.Equal(t, uuid.MustParse("123e4567-e89b-12d3-a456-426614174001"), deviceUUID)
}

func TestLogin(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Authorization", "Bearer faketoken123")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	token, err := Login(context.Background(), client, "user", "pass", uuid.New())

	assert.NoError(t, err)
	assert.Equal(t, "faketoken123", token) // теперь без "Bearer"
}

func TestLogin_NoAuthHeader(t *testing.T) {
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

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthClient_Register(t *testing.T) {
	expectedUUID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/register" {
			t.Errorf("unexpected URL path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedUUID.String()))
	}))
	defer server.Close()

	c := NewAuthClient(server.URL)
	userUUID, err := c.Register(context.Background(), "user1", "pass123")

	assert.NoError(t, err)
	assert.Equal(t, expectedUUID, userUUID)
}

func TestAuthClient_AddDevice(t *testing.T) {
	expectedUUID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/device/add" {
			t.Errorf("unexpected URL path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedUUID.String()))
	}))
	defer server.Close()

	c := NewAuthClient(server.URL)
	deviceUUID, err := c.AddDevice(context.Background(), "user1", "pass123", "pubkey123")

	assert.NoError(t, err)
	assert.Equal(t, expectedUUID, deviceUUID)
}

func TestAuthClient_Login(t *testing.T) {
	expectedToken := "Bearer testtoken123"
	deviceUUID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/login" {
			t.Errorf("unexpected URL path: %s", r.URL.Path)
		}
		w.Header().Set("Authorization", expectedToken)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewAuthClient(server.URL)
	token, err := c.Login(context.Background(), "user1", "pass123", deviceUUID)

	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)
}

func TestAuthClient_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	c := NewAuthClient(server.URL)

	_, err := c.Register(context.Background(), "user1", "pass123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "register failed")

	_, err = c.AddDevice(context.Background(), "user1", "pass123", "pubkey123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "add device failed")

	_, err = c.Login(context.Background(), "user1", "pass123", uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "login failed")
}

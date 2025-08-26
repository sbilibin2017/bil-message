package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoomClient(t *testing.T) {
	roomUUID := uuid.New()
	memberUUID := uuid.New()
	token := "test-token"

	// httptest сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+token {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/rooms":
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(roomUUID.String()))
				return
			}
		case fmt.Sprintf("/rooms/%s", roomUUID.String()):
			if r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusOK)
				return
			}
		case fmt.Sprintf("/rooms/%s/%s", roomUUID.String(), memberUUID.String()):
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusOK)
				return
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	restyClient := resty.New()
	restyClient.SetBaseURL(ts.URL)
	roomClient := NewRoomClient(restyClient)
	ctx := context.Background()

	// CreateRoom
	createdRoomUUID, err := roomClient.CreateRoom(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, roomUUID, createdRoomUUID)

	// DeleteRoom
	err = roomClient.DeleteRoom(ctx, token, roomUUID)
	require.NoError(t, err)

	// AddMember
	err = roomClient.AddMember(ctx, token, roomUUID, memberUUID)
	require.NoError(t, err)

	// RemoveMember
	err = roomClient.RemoveMember(ctx, token, roomUUID, memberUUID)
	require.NoError(t, err)
}

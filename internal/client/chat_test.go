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
)

func TestCreateChat(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "123e4567-e89b-12d3-a456-426614174000")
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	roomUUID, err := CreateChat(context.Background(), client, "token123")

	assert.NoError(t, err)
	assert.Equal(t, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"), roomUUID)
}

func TestCreateChat_NetworkError(t *testing.T) {
	client := resty.New().SetBaseURL("http://127.0.0.1:0")
	roomUUID, err := CreateChat(context.Background(), client, "token123")

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, roomUUID)
}

func TestCreateChat_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	roomUUID, err := CreateChat(context.Background(), client, "token123")

	assert.Error(t, err)
	assert.Equal(t, uuid.Nil, roomUUID)
}

func TestRemoveChat(t *testing.T) {
	roomUUID := uuid.New()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/"+roomUUID.String() {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	err := RemoveChat(context.Background(), client, "token123", roomUUID)

	assert.NoError(t, err)
}

func TestRemoveChat_NetworkError(t *testing.T) {
	roomUUID := uuid.New()
	client := resty.New().SetBaseURL("http://127.0.0.1:0")
	err := RemoveChat(context.Background(), client, "token123", roomUUID)

	assert.Error(t, err)
}

func TestRemoveChat_ServerError(t *testing.T) {
	roomUUID := uuid.New()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	err := RemoveChat(context.Background(), client, "token123", roomUUID)

	assert.Error(t, err)
}

func TestAddChatMember(t *testing.T) {
	roomUUID := uuid.New()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/"+roomUUID.String()+"/member" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	err := AddChatMember(context.Background(), client, "token123", roomUUID)

	assert.NoError(t, err)
}

func TestAddChatMember_NetworkError(t *testing.T) {
	roomUUID := uuid.New()
	client := resty.New().SetBaseURL("http://127.0.0.1:0")
	err := AddChatMember(context.Background(), client, "token123", roomUUID)

	assert.Error(t, err)
}

func TestAddChatMember_ServerError(t *testing.T) {
	roomUUID := uuid.New()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	err := AddChatMember(context.Background(), client, "token123", roomUUID)

	assert.Error(t, err)
}

func TestRemoveChatMember(t *testing.T) {
	roomUUID := uuid.New()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/"+roomUUID.String()+"/member" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	err := RemoveChatMember(context.Background(), client, "token123", roomUUID)

	assert.NoError(t, err)
}

func TestRemoveChatMember_NetworkError(t *testing.T) {
	roomUUID := uuid.New()
	client := resty.New().SetBaseURL("http://127.0.0.1:0")
	err := RemoveChatMember(context.Background(), client, "token123", roomUUID)

	assert.Error(t, err)
}

func TestRemoveChatMember_ServerError(t *testing.T) {
	roomUUID := uuid.New()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)
	err := RemoveChatMember(context.Background(), client, "token123", roomUUID)

	assert.Error(t, err)
}

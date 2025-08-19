package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRegisterer := NewMockRegisterer(ctrl)

	t.Run("successful registration", func(t *testing.T) {
		reqBody := []byte(`{"username":"user1","password":"Secret123!"}`)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		mockRegisterer.EXPECT().
			Register(gomock.Any(), "user1", "Secret123!").
			Return("token123", nil)

		handler := RegisterHandler(mockRegisterer)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Bearer token123", w.Header().Get("Authorization"))
	})

	t.Run("user already exists", func(t *testing.T) {
		reqBody := []byte(`{"username":"user2","password":"Secret123!"}`)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		mockRegisterer.EXPECT().
			Register(gomock.Any(), "user2", "Secret123!").
			Return("", services.ErrUserAlreadyExists)

		handler := RegisterHandler(mockRegisterer)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("bad request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte("not-json")))
		w := httptest.NewRecorder()

		handler := RegisterHandler(mockRegisterer)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		reqBody := []byte(`{"username":"user3","password":"Secret123!"}`)
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		mockRegisterer.EXPECT().
			Register(gomock.Any(), "user3", "Secret123!").
			Return("", errors.New("some error"))

		handler := RegisterHandler(mockRegisterer)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

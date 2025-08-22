package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/bil-message/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockReg := NewMockRegisterer(ctrl)
	handler := RegisterHandler(mockReg)

	t.Run("success", func(t *testing.T) {
		reqBody, _ := json.Marshal(RegisterRequest{
			Username: "user1",
			Password: "pass123",
		})

		mockReg.EXPECT().
			Register(gomock.Any(), "user1", "pass123").
			Return("test-token", nil)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "Bearer test-token", resp.Header.Get("Authorization"))
	})

	t.Run("bad request - invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte("{invalid-json")))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("conflict - user already exists", func(t *testing.T) {
		reqBody, _ := json.Marshal(RegisterRequest{
			Username: "user1",
			Password: "pass123",
		})

		mockReg.EXPECT().
			Register(gomock.Any(), "user1", "pass123").
			Return("", errors.ErrUserAlreadyExists)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		reqBody, _ := json.Marshal(RegisterRequest{
			Username: "user1",
			Password: "pass123",
		})

		mockReg.EXPECT().
			Register(gomock.Any(), "user1", "pass123").
			Return("", context.DeadlineExceeded)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestLoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLoginer := NewMockLoginer(ctrl)
	handler := LoginHandler(mockLoginer)

	t.Run("success", func(t *testing.T) {
		reqBody, _ := json.Marshal(LoginRequest{
			Username: "user1",
			Password: "pass123",
		})

		mockLoginer.EXPECT().
			Login(gomock.Any(), "user1", "pass123").
			Return("test-token", nil)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "Bearer test-token", resp.Header.Get("Authorization"))
	})

	t.Run("bad request - invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("{invalid-json")))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized - user not found", func(t *testing.T) {
		reqBody, _ := json.Marshal(LoginRequest{
			Username: "user1",
			Password: "pass123",
		})

		mockLoginer.EXPECT().
			Login(gomock.Any(), "user1", "pass123").
			Return("", errors.ErrUserNotFound)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		reqBody, _ := json.Marshal(LoginRequest{
			Username: "user1",
			Password: "pass123",
		})

		mockLoginer.EXPECT().
			Login(gomock.Any(), "user1", "pass123").
			Return("", context.DeadlineExceeded)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(reqBody))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

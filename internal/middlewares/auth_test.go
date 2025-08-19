package middlewares

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockParser := NewMockTokenParser(ctrl)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("missing token", func(t *testing.T) {
		mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("", errors.New("no token"))

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler := AuthMiddleware(mockParser)(nextHandler)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
		mockParser.EXPECT().Parse("token").Return(nil, errors.New("invalid"))

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler := AuthMiddleware(mockParser)(nextHandler)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("context error", func(t *testing.T) {
		payload := &models.TokenPayload{}
		mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
		mockParser.EXPECT().Parse("token").Return(payload, nil)
		mockParser.EXPECT().SetToContext(gomock.Any(), payload).Return(context.Background(), errors.New("context error"))

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler := AuthMiddleware(mockParser)(nextHandler)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("successful request", func(t *testing.T) {
		payload := &models.TokenPayload{}
		ctx := context.WithValue(context.Background(), "user", payload)

		mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
		mockParser.EXPECT().Parse("token").Return(payload, nil)
		mockParser.EXPECT().SetToContext(gomock.Any(), payload).Return(ctx, nil)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler := AuthMiddleware(mockParser)(nextHandler)
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

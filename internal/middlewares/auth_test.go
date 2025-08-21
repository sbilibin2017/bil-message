package middlewares

import (
	"context"
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

	nextHandlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true

		// Проверяем, что контекст содержит payload после SetToContext
		payload := r.Context().Value("payload").(*models.TokenPayload)
		assert.Equal(t, "11111111-1111-1111-1111-111111111111", payload.UserUUID)
		assert.Equal(t, "22222222-2222-2222-2222-222222222222", payload.DeviceUUID)
	})

	middleware := AuthMiddleware(mockParser)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		mockParser.EXPECT().GetFromRequest(req).Return(ptrString("token"), nil)
		mockParser.EXPECT().Parse("token").Return(&models.TokenPayload{
			UserUUID:   "11111111-1111-1111-1111-111111111111",
			DeviceUUID: "22222222-2222-2222-2222-222222222222",
		}, nil)
		mockParser.EXPECT().SetToContext(req.Context(), gomock.Any()).DoAndReturn(
			func(ctx context.Context, payload *models.TokenPayload) context.Context {
				return context.WithValue(ctx, "payload", payload)
			},
		)

		rr := httptest.NewRecorder()
		middleware(nextHandler).ServeHTTP(rr, req)

		assert.True(t, nextHandlerCalled)
	})

	t.Run("GetFromRequest fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		mockParser.EXPECT().GetFromRequest(req).Return(nil, assert.AnError)

		rr := httptest.NewRecorder()
		middleware(nextHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Parse fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		mockParser.EXPECT().GetFromRequest(req).Return(ptrString("token"), nil)
		mockParser.EXPECT().Parse("token").Return(nil, assert.AnError)

		rr := httptest.NewRecorder()
		middleware(nextHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

// helper to return pointer to string
func ptrString(s string) *string {
	return &s
}

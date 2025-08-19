package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockParser := NewMockTokenParser(ctrl)

	nextHandlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true

		// Проверяем, что контекст содержит UUID после SetToContext
		userUUID, clientUUID := r.Context().Value("user").([2]uuid.UUID)[0], r.Context().Value("user").([2]uuid.UUID)[1]
		assert.Equal(t, uuid.MustParse("11111111-1111-1111-1111-111111111111"), userUUID)
		assert.Equal(t, uuid.MustParse("22222222-2222-2222-2222-222222222222"), clientUUID)
	})

	middleware := AuthMiddleware(mockParser)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		mockParser.EXPECT().GetFromRequest(req).Return("token", nil)
		mockParser.EXPECT().Parse("token").Return(
			uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			uuid.MustParse("22222222-2222-2222-2222-222222222222"),
			nil,
		)
		mockParser.EXPECT().SetToContext(req.Context(),
			uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		).Return(context.WithValue(req.Context(), "user", [2]uuid.UUID{
			uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			uuid.MustParse("22222222-2222-2222-2222-222222222222"),
		}))

		rr := httptest.NewRecorder()
		middleware(nextHandler).ServeHTTP(rr, req)

		assert.True(t, nextHandlerCalled)
	})

	t.Run("GetFromRequest fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		mockParser.EXPECT().GetFromRequest(req).Return("", assert.AnError)

		rr := httptest.NewRecorder()
		middleware(nextHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("Parse fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		mockParser.EXPECT().GetFromRequest(req).Return("token", nil)
		mockParser.EXPECT().Parse("token").Return(uuid.Nil, uuid.Nil, assert.AnError)

		rr := httptest.NewRecorder()
		middleware(nextHandler).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

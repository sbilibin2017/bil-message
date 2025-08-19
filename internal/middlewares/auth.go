package middlewares

import (
	"context"
	"net/http"

	"github.com/sbilibin2017/bil-message/internal/models"
)

// TokenParser интерфейс для разбора и проверки токена
type TokenParser interface {
	GetFromRequest(r *http.Request) (string, error)
	Parse(tokenString string) (*models.TokenPayload, error)
	SetToContext(ctx context.Context, payload *models.TokenPayload) (context.Context, error)
}

// AuthMiddleware возвращает middleware, использующее TokenParser
func AuthMiddleware(parser TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, err := parser.GetFromRequest(r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			payload, err := parser.Parse(tokenString)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx, err := parser.SetToContext(r.Context(), payload)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

package middlewares

import (
	"context"
	"net/http"

	"github.com/sbilibin2017/bil-message/internal/models"
)

// TokenParser интерфейс для работы с JWT
type TokenParser interface {
	GetFromRequest(r *http.Request) (*string, error)
	Parse(tokenString string) (*models.TokenPayload, error)
	SetToContext(ctx context.Context, payload *models.TokenPayload) context.Context
}

// AuthMiddleware возвращает middleware, использующее TokenParser
func AuthMiddleware(parser TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем токен из запроса
			tokenString, err := parser.GetFromRequest(r)
			if err != nil || tokenString == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Парсим токен
			payload, err := parser.Parse(*tokenString)
			if err != nil || payload == nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Сохраняем в контекст
			ctx := parser.SetToContext(r.Context(), payload)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

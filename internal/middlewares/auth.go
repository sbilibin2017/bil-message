package middlewares

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// TokenParser интерфейс для работы с JWT
type TokenParser interface {
	GetFromRequest(r *http.Request) (string, error)
	Parse(tokenString string) (uuid.UUID, uuid.UUID, error)
	SetToContext(ctx context.Context, userUUID uuid.UUID, clientUUID uuid.UUID) context.Context
}

// AuthMiddleware возвращает middleware, использующее TokenParser
func AuthMiddleware(parser TokenParser) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Извлекаем токен из запроса
			tokenString, err := parser.GetFromRequest(r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Парсим токен
			userUUID, clientUUID, err := parser.Parse(tokenString)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Сохраняем в контекст
			ctx := parser.SetToContext(r.Context(), userUUID, clientUUID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

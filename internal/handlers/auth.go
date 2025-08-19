package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/services"
)

// RegisterRequest определяет входящий запрос на регистрацию пользователя.
// swagger:model RegisterRequest
type RegisterRequest struct {
	// Имя пользователя
	// required: true
	// example: johndoe
	// default: user123
	Username string `json:"username"`
	// Пароль пользователя
	// required: true
	// example: Secret123!
	// default: P@ssw0rd
	Password string `json:"password"`
}

// Registerer определяет интерфейс для регистрации пользователя и получения JWT токена.
type Registerer interface {
	Register(ctx context.Context, username string, password string) (userUUID uuid.UUID, err error)
}

// RegisterHandler обрабатывает регистрацию пользователя и возвращает токен в заголовке.
// @Summary      Регистрация нового пользователя
// @Description  Создает новый аккаунт пользователя. JWT токен возвращается в заголовке Authorization: Bearer <token>.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body handlers.RegisterRequest true "Запрос на регистрацию пользователя"
// @Success      200 {object} map[string]string "Успешная регистрация, токен в заголовке Authorization"
// @Failure      400 {string} string "Неверный запрос или невалидные данные"
// @Failure      409 {string} string "Пользователь с таким именем уже существует"
// @Failure      500 {string} string "Внутренняя ошибка сервера"
// @Router       /auth/register [post]
func RegisterHandler(
	reg Registerer,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest

		// Декодируем JSON-запрос
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Вызываем регистрацию
		userUUID, err := reg.Register(r.Context(), req.Username, req.Password)
		if err != nil {
			if errors.Is(err, services.ErrUserAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Устанавливаем Content-Type и возвращаем UUID в plain text
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userUUID.String()))
	}
}

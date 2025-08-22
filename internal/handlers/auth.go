package handlers

import (
	"context"
	"encoding/json"
	"log"

	"net/http"

	"github.com/sbilibin2017/bil-message/internal/errors"
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
	Register(ctx context.Context, username, password string) (tokenString string, err error)
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
		token, err := reg.Register(r.Context(), req.Username, req.Password)
		if err != nil {
			switch err {
			case errors.ErrUserAlreadyExists:
				w.WriteHeader(http.StatusConflict)
				return
			default:
				log.Printf("op: %s, err:%s", "user register", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		}

		// Устанавливаем Content-Type и возвращаем UUID в plain text
		w.Header().Set("Authorization", "Bearer "+token)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}

// LoginRequest определяет входящий запрос на аутентификацию пользователя.
// swagger:model LoginRequest
type LoginRequest struct {
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

// LoginRequest определяет интерфейс для регистрации пользователя и получения JWT токена.
type Loginer interface {
	Login(ctx context.Context, username, password string) (tokenString string, err error)
}

// LoginHandler обрабатывает аутентификацию пользователя и возвращает JWT токен в заголовке Authorization.
// @Summary      Вход пользователя
// @Description  Проверяет пользователя и устройство, возвращает JWT токен в заголовке Authorization: Bearer <token>.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request body handlers.LoginRequest true "Запрос на вход пользователя"
// @Success      200 {string} string "Успешная аутентификация, токен в заголовке Authorization"
// @Failure      400 {string} string "Неверный запрос или невалидные данные"
// @Failure      401 {string} string "Пользователь не найден, неверный пароль или устройство не найдено"
// @Failure      500 {string} string "Внутренняя ошибка сервера"
// @Router       /auth/login [post]
func LoginHandler(
	loginer Loginer,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest

		// Декодируем JSON-запрос
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Вызываем метод логина
		token, err := loginer.Login(r.Context(), req.Username, req.Password)
		if err != nil {
			log.Printf("op: %s, err:%s", "user login", err.Error())
			switch err {
			case errors.ErrUserNotFound, errors.ErrInvalidPassword, errors.ErrDeviceNotFound:
				w.WriteHeader(http.StatusUnauthorized) // 401 для ошибок аутентификации
				return
			default:
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		// Устанавливаем токен в заголовок Authorization
		w.Header().Set("Authorization", "Bearer "+token)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}

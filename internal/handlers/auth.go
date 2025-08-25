package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/services"
)

type Registerer interface {
	Register(ctx context.Context, username string, password string) (userUUID uuid.UUID, err error)
}

// RegisterRequest представляет JSON тело запроса на регистрацию.
// swagger:model RegisterRequest
type RegisterRequest struct {
	// Username пользователя
	// required: true
	// example: johndoe
	Username string `json:"username"`

	// Пароль пользователя
	// required: true
	// example: mySecret123
	Password string `json:"password"`
}

// RegisterHandler
// @Summary Регистрация нового пользователя
// @Description Создаёт нового пользователя с заданными username и password
// @Tags Auth
// @Accept json
// @Produce plain
// @Param request body RegisterRequest true "Данные пользователя"
// @Success 200 "Пользователь успешно зарегистрирован"
// @Failure 400 "Некорректные данные запроса"
// @Failure 409 "Пользователь с таким именем уже существует"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /auth/register [post]
func RegisterHandler(svc Registerer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Username == "" || req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userUUID, err := svc.Register(r.Context(), req.Username, req.Password)
		if err != nil {
			if errors.Is(err, services.ErrUsernameAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userUUID.String()))
	}
}

type DeviceAdder interface {
	AddDevice(ctx context.Context, username, password, publicKey string) (deviceUUID uuid.UUID, err error)
}

// DeviceRequest представляет JSON тело запроса на добавление устройства.
// swagger:model DeviceRequest
type DeviceRequest struct {
	// Username пользователя
	// required: true
	// example: johndoe
	Username string `json:"username"`

	// Пароль пользователя
	// required: true
	// example: mySecret123
	Password string `json:"password"`

	// Публичный ключ устройства
	// required: true
	// example: MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAr...
	PublicKey string `json:"public_key"`
}

// AddDeviceHandler
// @Summary Добавление нового устройства
// @Description Привязывает новое устройство к пользователю и возвращает UUID устройства
// @Tags Auth
// @Accept json
// @Produce plain
// @Param request body DeviceRequest true "Данные устройства"
// @Success 200 {string} string "UUID устройства"
// @Failure 400 "Неверные учетные данные или некорректные данные запроса"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /auth/device [post]
func AddDeviceHandler(svc DeviceAdder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DeviceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Username == "" || req.Password == "" || req.PublicKey == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		deviceUUID, err := svc.AddDevice(r.Context(), req.Username, req.Password, req.PublicKey)
		if err != nil {
			if errors.Is(err, services.ErrInvalidCredentials) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte(deviceUUID.String()))
	}
}

type Loginer interface {
	Login(ctx context.Context, username, password string, deviceUUID uuid.UUID) (token string, err error)
}

// LoginRequest представляет JSON тело запроса на логин.
// swagger:model LoginRequest
type LoginRequest struct {
	// Username пользователя
	// required: true
	// example: johndoe
	Username string `json:"username"`

	// Пароль пользователя
	// required: true
	// example: mySecret123
	Password string `json:"password"`

	// UUID устройства
	// required: true
	// example: 4c2b87cd-8c44-4546-89df-7a751cbac96e
	DeviceUUID string `json:"device_uuid"`
}

// LoginHandler
// @Summary Вход пользователя
// @Description Проверяет username, password и deviceUUID, возвращает JWT в заголовке Authorization
// @Tags Auth
// @Accept json
// @Produce plain
// @Param request body LoginRequest true "Данные для входа"
// @Success 200 "JWT токен успешно сгенерирован и возвращен в заголовке Authorization"
// @Failure 400 "Неверные учетные данные или некорректные данные запроса"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func LoginHandler(svc Loginer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Username == "" || req.Password == "" || req.DeviceUUID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		deviceUUID, err := uuid.Parse(req.DeviceUUID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := svc.Login(r.Context(), req.Username, req.Password, deviceUUID)
		if err != nil {
			if errors.Is(err, services.ErrInvalidCredentials) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	}
}

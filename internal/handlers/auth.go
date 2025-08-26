package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/services"
)

// Registerer интерфейс для регистрации пользователя
type Registerer interface {
	Register(ctx context.Context, username, password string) (userUUID uuid.UUID, err error)
}

// RegisterRequest представляет JSON тело запроса на регистрацию пользователя
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

// NewRegisterHandler
// @Summary Регистрация нового пользователя
// @Description Создаёт нового пользователя с заданными username и password
// @Tags Auth
// @Accept json
// @Produce plain
// @Param request body RegisterRequest true "Данные пользователя"
// @Success 200 {string} string "UUID нового пользователя"
// @Failure 400 "Некорректные данные запроса"
// @Failure 409 "Пользователь с таким именем уже существует"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /auth/register [post]
func NewRegisterHandler(svc Registerer) http.HandlerFunc {
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
			switch err {
			case services.ErrUserExists:
				w.WriteHeader(http.StatusConflict)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userUUID.String()))
	}
}

// DeviceAdder интерфейс для добавления устройства пользователю
type DeviceAdder interface {
	AddDevice(ctx context.Context, username, password, publicKey string) (deviceUUID uuid.UUID, err error)
}

// AddDeviceRequest представляет JSON тело запроса на добавление устройства
// swagger:model AddDeviceRequest
type AddDeviceRequest struct {
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

// NewDeviceAddHandler
// @Summary Добавление нового устройства
// @Description Привязывает новое устройство к пользователю и возвращает UUID устройства
// @Tags Auth
// @Accept json
// @Produce plain
// @Param request body AddDeviceRequest true "Данные устройства"
// @Success 200 {string} string "UUID устройства"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неверные учетные данные"
// @Failure 404 "Пользователь не найден"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /auth/device [post]
func NewDeviceAddHandler(svc DeviceAdder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AddDeviceRequest
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
			switch err {
			case services.ErrUserNotFound:
				w.WriteHeader(http.StatusNotFound)
			case services.ErrInvalidCredential:
				w.WriteHeader(http.StatusUnauthorized)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(deviceUUID.String()))
	}
}

// Loginer интерфейс для входа пользователя и получения JWT
type Loginer interface {
	Login(ctx context.Context, username, password string, deviceUUID uuid.UUID) (tokenString string, err error)
}

// LoginRequest представляет JSON тело запроса на логин
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

// NewLoginHandler
// @Summary Вход пользователя
// @Description Проверяет username, password и deviceUUID, возвращает JWT в заголовке Authorization
// @Tags Auth
// @Accept json
// @Produce plain
// @Param request body LoginRequest true "Данные для входа"
// @Success 200 {string} string "JWT токен возвращается в заголовке Authorization"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неверные учетные данные"
// @Failure 404 "Пользователь или устройство не найдены"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /auth/login [post]
func NewLoginHandler(svc Loginer) http.HandlerFunc {
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
			switch err {
			case services.ErrUserNotFound, services.ErrDeviceNotFound:
				w.WriteHeader(http.StatusNotFound)
			case services.ErrInvalidCredential:
				w.WriteHeader(http.StatusUnauthorized)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	}
}

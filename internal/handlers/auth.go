package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/services"
)

//
// Register Handler
//

type Registerer interface {
	Register(ctx context.Context, username string, password string) error
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
			log.Printf("[RegisterHandler] Ошибка декодирования запроса: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Username == "" || req.Password == "" {
			log.Printf("[RegisterHandler] Пустое имя пользователя или пароль")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := svc.Register(r.Context(), req.Username, req.Password); err != nil {
			if errors.Is(err, services.ErrUsernameAlreadyExists) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			log.Printf("[RegisterHandler] Ошибка регистрации пользователя '%s': %v", req.Username, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

//
// Add Device Handler
//

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
// @Tags Device
// @Accept json
// @Produce plain
// @Param request body DeviceRequest true "Данные устройства"
// @Success 200 {string} string "UUID устройства"
// @Failure 400 "Неверные учетные данные или некорректные данные запроса"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /device [post]
func AddDeviceHandler(svc DeviceAdder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DeviceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("[AddDeviceHandler] Ошибка декодирования запроса: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Username == "" || req.Password == "" || req.PublicKey == "" {
			log.Printf("[AddDeviceHandler] Пустое имя пользователя, пароль или публичный ключ")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		deviceUUID, err := svc.AddDevice(r.Context(), req.Username, req.Password, req.PublicKey)
		if err != nil {
			if errors.Is(err, services.ErrInvalidCredentials) {
				log.Printf("[AddDeviceHandler] Неверные учетные данные для пользователя '%s'", req.Username)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			log.Printf("[AddDeviceHandler] Ошибка добавления устройства для пользователя '%s': %v", req.Username, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("[AddDeviceHandler] Устройство для пользователя '%s' успешно добавлено: %s", req.Username, deviceUUID)
		w.Write([]byte(deviceUUID.String()))
	}
}

//
// Login Handler
//

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
			log.Printf("[LoginHandler] Ошибка декодирования запроса: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Username == "" || req.Password == "" || req.DeviceUUID == "" {
			log.Printf("[LoginHandler] Пустое имя пользователя, пароль или UUID устройства")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		deviceUUID, err := uuid.Parse(req.DeviceUUID)
		if err != nil {
			log.Printf("[LoginHandler] Некорректный UUID устройства: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := svc.Login(r.Context(), req.Username, req.Password, deviceUUID)
		if err != nil {
			if errors.Is(err, services.ErrInvalidCredentials) {
				log.Printf("[LoginHandler] Неверные учетные данные для пользователя '%s'", req.Username)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			log.Printf("[LoginHandler] Ошибка входа для пользователя '%s': %v", req.Username, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("[LoginHandler] Пользователь '%s' успешно вошёл. JWT: %s", req.Username, token)
		w.Header().Set("Authorization", "Bearer "+token)
		w.WriteHeader(http.StatusOK)
	}
}

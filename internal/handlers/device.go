package handlers

import (
	"context"
	"encoding/json"

	"net/http"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/configs/log"
	"github.com/sbilibin2017/bil-message/internal/errors"
)

// DeviceRequest определяет входящий запрос на регистрацию устройства.
// swagger:model DeviceRequest
type DeviceRequest struct {
	// UUID пользователя
	// required: true
	// example: 550e8400-e29b-41d4-a716-446655440000
	UserUUID string `json:"user_uuid"`
	// Публичный ключ устройства
	// required: true
	// example: ssh-rsa AAAAB3Nza...
	PublicKey string `json:"public_key"`
}

// DeviceRegisterer определяет интерфейс для регистрации устройства.
type DeviceRegisterer interface {
	Register(ctx context.Context, userUUID uuid.UUID, publicKey string) (*uuid.UUID, error)
}

// DeviceRegisterHandler обрабатывает регистрацию устройства и возвращает deviceUUID.
// @Summary      Регистрация нового устройства
// @Description  Привязывает устройство к пользователю по userUUID.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        request body handlers.DeviceRequest true "Запрос на регистрацию устройства"
// @Success      200 {string} string "Успешная регистрация устройства, возвращается deviceUUID"
// @Failure      400 {string} string "Неверный запрос или невалидные данные"
// @Failure      404 {string} string "Пользователь не найден"
// @Failure      500 {string} string "Внутренняя ошибка сервера"
// @Router       /devices/register [post]
func DeviceRegisterHandler(
	reg DeviceRegisterer,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req DeviceRequest

		// Декодируем JSON-запрос
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Парсим userUUID
		userUUID, err := uuid.Parse(req.UserUUID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Вызываем регистрацию
		deviceUUID, err := reg.Register(r.Context(), userUUID, req.PublicKey)
		if err != nil {
			if err == errors.ErrUserNotFound {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			log.Log("device register", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Устанавливаем Content-Type и возвращаем UUID устройства
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(deviceUUID.String()))
	}
}

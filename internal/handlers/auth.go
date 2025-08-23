package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
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

// RegisterHandler godoc
// @Summary Регистрация нового пользователя
// @Description Создаёт нового пользователя с заданными username и password
// @Tags auth
// @Accept json
// @Produce plain
// @Param request body RegisterRequest true "Данные пользователя"
// @Success 200
// @Failure 400
// @Failure 500
// @Router /auth/register [post]
func RegisterHandler(svc Registerer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Username == "" || req.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userUUID, err := svc.Register(r.Context(), req.Username, req.Password)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userUUID.String()))
	}
}

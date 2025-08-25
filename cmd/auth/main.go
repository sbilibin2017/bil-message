package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var jwtSecret = []byte("supersecretkey")

type User struct {
	UUID    string
	Devices []string
}

var users = map[string]*User{}

type Claims struct {
	UserUUID   string `json:"user_uuid"`
	DeviceUUID string `json:"device_uuid"`
	jwt.RegisteredClaims
}

// Регистрация пользователя
func registerHandler(w http.ResponseWriter, r *http.Request) {
	userUUID := uuid.NewString()
	users[userUUID] = &User{UUID: userUUID}
	json.NewEncoder(w).Encode(map[string]string{"user_uuid": userUUID})
}

// Добавление устройства
func addDeviceHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserUUID string `json:"user_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	user, ok := users[req.UserUUID]
	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	deviceUUID := uuid.NewString()
	user.Devices = append(user.Devices, deviceUUID)
	json.NewEncoder(w).Encode(map[string]string{"device_uuid": deviceUUID})
}

// Аутентификация пользователя — выдача токена
func tokenHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserUUID   string `json:"user_uuid"`
		DeviceUUID string `json:"device_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	user, ok := users[req.UserUUID]
	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	valid := false
	for _, d := range user.Devices {
		if d == req.DeviceUUID {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, "device not registered", http.StatusUnauthorized)
		return
	}

	claims := Claims{
		UserUUID:   req.UserUUID,
		DeviceUUID: req.DeviceUUID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString(jwtSecret)
	json.NewEncoder(w).Encode(map[string]string{"token": signed})
}

// Декодирование токена с проверкой подписи и срока действия
func tokenDecodeHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	t, err := jwt.ParseWithClaims(req.Token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	claims, ok := t.Claims.(*Claims)
	if !ok || !t.Valid {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"user_uuid":   claims.UserUUID,
		"device_uuid": claims.DeviceUUID,
	})
}

func main() {
	r := chi.NewRouter()

	// Префикс /api/v1/auth для всех эндпоинтов
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", registerHandler)        // регистрация пользователя
		r.Post("/device/add", addDeviceHandler)     // добавление устройства
		r.Post("/token", tokenHandler)              // выдача токена
		r.Post("/token/decode", tokenDecodeHandler) // декодирование токена
	})

	log.Println("Auth server running on :9000")
	log.Fatal(http.ListenAndServe(":9000", r))
}

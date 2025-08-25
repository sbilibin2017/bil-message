package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/sbilibin2017/bil-message/internal/facades"
	"github.com/sbilibin2017/bil-message/internal/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ChatHandler возвращает функцию-обработчик для WebSocket,
// принимает authCheck closure и NATS connection
func ChatHandler(authFacade *facades.AuthFacade, nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomUUID := chi.URLParam(r, "room-uuid")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := authFacade.Decode(r.Context(), token)
		if err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade error:", err)
			return
		}
		defer conn.Close()

		log.Printf("User %s connected to room %s", claims.UserUUID, roomUUID)

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("ReadMessage error:", err)
				break
			}

			m := models.Message{
				RoomUUID:   roomUUID,
				UserUUID:   claims.UserUUID,
				DeviceUUID: claims.DeviceUUID,
				Text:       string(msg),
				Timestamp:  time.Now().UnixMilli(),
			}
			data, _ := json.Marshal(m)

			if err := nc.Publish("messages", data); err != nil {
				log.Println("Failed to publish message:", err)
			}
		}

		log.Printf("User %s disconnected from room %s", claims.UserUUID, roomUUID)
	}
}

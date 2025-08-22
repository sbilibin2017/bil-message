package handlers

import (
	"context"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

// JWTParser отвечает за работу с JWT-токенами.
type JWTParser interface {
	GetFromRequest(r *http.Request) (string, error)
	GetUserUUID(tokenString string) (string, error)
}

// ChatCreator интерфейс для создания чатов.
type ChatCreator interface {
	CreateChat(
		ctx context.Context,
		createdByUUID string,
	) (chatUUID string, err error)
}

// ChatMemberAdder интерфейс для добавления участников в чат.
type ChatMemberAdder interface {
	AddMember(
		ctx context.Context,
		chatUUID string,
		participantUUID string,
	) error
}

// NewCreateChatHandler создаёт HTTP-обработчик для создания нового чата.
func NewCreateChatHandler(jwtParser JWTParser, svc ChatCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := jwtParser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userUUID, err := jwtParser.GetUserUUID(token)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		chatUUID, err := svc.CreateChat(r.Context(), userUUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(chatUUID))
	}
}

// NewAddMemberHandler создаёт HTTP-обработчик для добавления пользователя в чат.
func NewAddMemberHandler(jwtParser JWTParser, svc ChatMemberAdder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := jwtParser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userUUID, err := jwtParser.GetUserUUID(token)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		chatUUID := chi.URLParam(r, "chat-uuid")
		if chatUUID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := svc.AddMember(r.Context(), chatUUID, userUUID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// ChatReader интерфейс для чтения информации о чате
type ChatReader interface {
	IsMember(
		ctx context.Context,
		chatUUID string,
		userUUID string,
	) (bool, error)
}

// upgrader настраивает WebSocket соединение.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// NewChatWSHandler создаёт HTTP-обработчик для WebSocket чата с проверкой участников.
func NewChatWSHandler(jwtParser JWTParser, chatReader ChatReader) http.HandlerFunc {
	rooms := make(map[string]map[string]*websocket.Conn)
	var roomsMux sync.RWMutex

	return func(w http.ResponseWriter, r *http.Request) {
		// JWT
		token, err := jwtParser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userUUID, err := jwtParser.GetUserUUID(token)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Chat UUID
		chatUUID := chi.URLParam(r, "chat-uuid")
		if chatUUID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Проверяем, что пользователь действительно участник чата
		isMember, err := chatReader.IsMember(r.Context(), chatUUID, userUUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !isMember {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Апгрейд соединения до WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() {
			conn.Close()
			roomsMux.Lock()
			if conns, ok := rooms[chatUUID]; ok {
				delete(conns, userUUID)
				if len(conns) == 0 {
					delete(rooms, chatUUID)
				}
			}
			roomsMux.Unlock()
		}()

		// Добавляем пользователя в комнату
		roomsMux.Lock()
		if _, ok := rooms[chatUUID]; !ok {
			rooms[chatUUID] = make(map[string]*websocket.Conn)
		}
		rooms[chatUUID][userUUID] = conn
		roomsMux.Unlock()

		// Основной цикл получения и пересылки сообщений
		for {
			var msg struct {
				Message string `json:"message"`
			}
			if err := conn.ReadJSON(&msg); err != nil {
				break
			}

			roomsMux.RLock()
			for id, c := range rooms[chatUUID] {
				if id != userUUID {
					_ = c.WriteJSON(msg)
				}
			}
			roomsMux.RUnlock()
		}
	}
}

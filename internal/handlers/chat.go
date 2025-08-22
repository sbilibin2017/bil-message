package handlers

import (
	"context"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// JWTParser отвечает за работу с JWT-токенами.
type JWTParser interface {
	// GetFromRequest извлекает токен из HTTP-запроса.
	GetFromRequest(r *http.Request) (*string, error)
	// Parse проверяет токен и возвращает UUID пользователя и клиента.
	Parse(tokenString string) (*models.TokenPayload, error)
}

// ChatCreator интерфейс для создания чатов.
type ChatCreator interface {
	// CreateChat создаёт новый чат с создателем creatorUUID.
	CreateChat(ctx context.Context, userUUID string) (*string, error)
}

// ChatMemberAdder интерфейс для добавления участников в чат.
type ChatMemberAdder interface {
	// AddMember добавляет пользователя userUUID в чат chatUUID.
	AddMember(ctx context.Context, chatUUID string, userUUID string) error
}

// NewCreateChatHandler создаёт HTTP-обработчик для создания нового чата.
func NewCreateChatHandler(jwtParser JWTParser, svc ChatCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := jwtParser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		payload, err := jwtParser.Parse(*token)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		chatUUID, err := svc.CreateChat(r.Context(), payload.UserUUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(*chatUUID))
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

		if _, err := jwtParser.Parse(*token); err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		chatUUID := chi.URLParam(r, "chat-uuid")
		userUUID := chi.URLParam(r, "user-uuid")

		if chatUUID == "" || userUUID == "" {
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

// upgrader настраивает WebSocket соединение.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// NewChatWSHandler создаёт HTTP-обработчик для WebSocket чата.
// Использует JWTParser для аутентификации пользователя и хранит подключения в памяти.
func NewChatWSHandler(jwtParser JWTParser) http.HandlerFunc {
	// Хранилище комнат: chatUUID -> map[userUUID]*websocket.Conn
	rooms := make(map[string]map[string]*websocket.Conn)
	var roomsMux sync.RWMutex

	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем токен
		token, err := jwtParser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Парсим токен
		payload, err := jwtParser.Parse(*token)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		// Получаем UUID чата из URL
		chatUUID := chi.URLParam(r, "chat-uuid")
		if chatUUID == "" {
			w.WriteHeader(http.StatusBadRequest)
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
			// Удаляем пользователя из комнаты при отключении
			roomsMux.Lock()
			if conns, ok := rooms[chatUUID]; ok {
				delete(conns, payload.UserUUID)
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
		rooms[chatUUID][payload.UserUUID] = conn
		roomsMux.Unlock()

		// Основной цикл чтения и пересылки сообщений
		for {
			var msg struct {
				Message string `json:"message"`
			}
			if err := conn.ReadJSON(&msg); err != nil {
				break
			}

			// Рассылаем сообщение всем участникам кроме отправителя
			roomsMux.RLock()
			for id, c := range rooms[chatUUID] {
				if id != payload.UserUUID {
					c.WriteJSON(msg)
				}
			}
			roomsMux.RUnlock()
		}
	}
}

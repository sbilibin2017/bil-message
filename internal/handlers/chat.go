package handlers

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sbilibin2017/bil-message/internal/chat"
	"github.com/sbilibin2017/bil-message/internal/jwt"
	"github.com/sbilibin2017/bil-message/internal/services"
)

// Интерфейсы для работы с комнатами и участниками
type RoomCreator interface {
	// CreateRoom создаёт новую комнату для пользователя и возвращает UUID комнаты
	CreateRoom(ctx context.Context, userUUID uuid.UUID) (roomUUID uuid.UUID, err error)
}

type RoomRemover interface {
	// RemoveRoom удаляет комнату по UUID
	RemoveRoom(ctx context.Context, roomUUID uuid.UUID) error
}

type RoomMemberAdder interface {
	// AddRoomMember добавляет пользователя в комнату
	AddRoomMember(ctx context.Context, roomUUID uuid.UUID, userUUID uuid.UUID) error
}

type RoomMemberRemover interface {
	// RemoveRoomMember удаляет пользователя из комнаты
	RemoveRoomMember(ctx context.Context, roomUUID uuid.UUID, userUUID uuid.UUID) error
}

type TokenParser interface {
	// GetFromRequest получает токен из HTTP-запроса
	GetFromRequest(r *http.Request) (tokenString string, err error)
	// Parse парсит токен и возвращает UUID пользователя и устройства
	Parse(tokenString string) (userUUID uuid.UUID, deviceUUID uuid.UUID, err error)
}

// CreateChatHandler создаёт новую комнату для текущего пользователя
// @Summary Создание новой комнаты
// @Description Создаёт новую комнату для текущего пользователя
// @Tags Chat
// @Accept plain
// @Produce plain
// @Success 200 {string} string "UUID комнаты"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неавторизован"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /chat [post]
func CreateChatHandler(svc RoomCreator, parser TokenParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := parser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userUUID, _, err := parser.Parse(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		roomUUID, err := svc.CreateRoom(r.Context(), userUUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write([]byte(roomUUID.String()))
	}
}

// RemoveChatHandler удаляет комнату по UUID
// @Summary Удаление комнаты
// @Description Удаляет комнату по UUID
// @Tags Chat
// @Accept plain
// @Produce plain
// @Param room-uuid path string true "UUID комнаты"
// @Success 200 "Комната успешно удалена"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неавторизован"
// @Failure 404 "Комната не найдена"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /chat/{room-uuid} [delete]
func RemoveChatHandler(svc RoomRemover, parser TokenParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, "room-uuid")

		roomUUID, err := uuid.Parse(roomID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := parser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _, err = parser.Parse(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := svc.RemoveRoom(r.Context(), roomUUID); err != nil {
			if errors.Is(err, services.ErrRoomNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// AddChatMemberHandler добавляет пользователя в комнату
// @Summary Добавление пользователя в комнату
// @Description Добавляет указанного пользователя (member-uuid) в комнату
// @Tags Chat
// @Accept plain
// @Produce plain
// @Param room-uuid path string true "UUID комнаты"
// @Param member-uuid path string true "UUID пользователя"
// @Success 200 "Пользователь успешно добавлен"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неавторизован"
// @Failure 404 "Комната не найдена"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /chat/{room-uuid}/{member-uuid} [post]
func AddChatMemberHandler(svc RoomMemberAdder, parser TokenParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, "room-uuid")
		memberID := chi.URLParam(r, "member-uuid")

		roomUUID, err := uuid.Parse(roomID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		memberUUID, err := uuid.Parse(memberID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := parser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _, err = parser.Parse(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := svc.AddRoomMember(r.Context(), roomUUID, memberUUID); err != nil {
			if errors.Is(err, services.ErrRoomNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// RemoveChatMemberHandler удаляет пользователя из комнаты
// @Summary Удаление пользователя из комнаты
// @Description Удаляет указанного пользователя (member-uuid) из комнаты
// @Tags Chat
// @Accept plain
// @Produce plain
// @Param room-uuid path string true "UUID комнаты"
// @Param member-uuid path string true "UUID пользователя"
// @Success 200 "Пользователь успешно удалён"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неавторизован"
// @Failure 404 "Комната или пользователь не найдены"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /chat/{room-uuid}/{member-uuid} [delete]
func RemoveChatMemberHandler(svc RoomMemberRemover, parser TokenParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, "room-uuid")
		memberID := chi.URLParam(r, "member-uuid")

		roomUUID, err := uuid.Parse(roomID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		memberUUID, err := uuid.Parse(memberID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		token, err := parser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _, err = parser.Parse(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := svc.RemoveRoomMember(r.Context(), roomUUID, memberUUID); err != nil {
			if errors.Is(err, services.ErrRoomNotFound) {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// ChatWebSocketHandler возвращает http.HandlerFunc для WebSocket соединений.
//
// Подключение по WebSocket осуществляется по пути /chat/ws/{room-uuid}.
// Пользователь должен быть авторизован и передать валидный токен в заголовке Authorization.
//
// После подключения клиент создается и добавляется в комнату.
// Если комната с заданным UUID не существует, она создается автоматически.
//
// Чтение и запись сообщений происходят асинхронно через ReadPump и WritePump.
//
// @Summary WebSocket соединение для чата
// @Description Создает WebSocket соединение для конкретной комнаты. Сообщения рассылаются всем участникам, кроме отправителя.
// @Tags Chat
// @Accept plain
// @Produce json
// @Param room-uuid path string true "UUID комнаты"
// @Success 101 "WebSocket соединение установлено"
// @Failure 400 "Некорректный UUID комнаты"
// @Failure 401 "Неавторизован"
// @Failure 500 "Ошибка сервера при апгрейде соединения"
// @Router /chat/{room-uuid}/ws [get]
func ChatWebSocketHandler(
	newClient func(conn *websocket.Conn, userUUID, roomUUID uuid.UUID) *chat.ChatClient,
	newRoom func(roomUUID uuid.UUID) *chat.ChatRoom,
	parser *jwt.JWT,
) http.HandlerFunc {

	rooms := make(map[uuid.UUID]*chat.ChatRoom)
	var mu sync.Mutex

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Получаем UUID комнаты из URL
		roomIDStr := chi.URLParam(r, "room-uuid")
		roomUUID, err := uuid.Parse(roomIDStr)
		if err != nil {
			http.Error(w, "invalid room UUID", http.StatusBadRequest)
			return
		}

		// Получаем токен из запроса
		token, err := parser.GetFromRequest(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Парсим токен
		userUUID, _, err := parser.Parse(token)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Апгрейдим соединение в WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "failed to upgrade websocket", http.StatusInternalServerError)
			return
		}

		client := newClient(conn, userUUID, roomUUID)

		// Получаем или создаём комнату
		mu.Lock()
		room, ok := rooms[roomUUID]
		if !ok {
			room = newRoom(roomUUID)
			rooms[roomUUID] = room
		}
		mu.Unlock()

		room.AddClient(client)

		// Запускаем неблокирующие горутины для чтения и записи
		client.ReadPump(room)
		client.WritePump()
	}
}

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/sbilibin2017/bil-message/internal/services"
)

// RoomCreator интерфейс для создания комнаты
type RoomCreator interface {
	CreateRoom(ctx context.Context, userUUID uuid.UUID) (uuid.UUID, error)
}

// RoomDeleter интерфейс для удаления комнаты
type RoomDeleter interface {
	DeleteRoom(ctx context.Context, userUUID, roomUUID uuid.UUID) error
}

// RoomMemberAdder интерфейс для добавления участника в комнату
type RoomMemberAdder interface {
	AddMember(ctx context.Context, userUUID, roomUUID, memberUUID uuid.UUID) error
}

// RoomMemberRemover интерфейс для удаления участника из комнаты
type RoomMemberRemover interface {
	RemoveMember(ctx context.Context, userUUID, roomUUID, memberUUID uuid.UUID) error
}

// TokenParser интерфейс для получения и парсинга JWT
type TokenParser interface {
	GetFromRequest(r *http.Request) (tokenString string, err error)
	Parse(tokenString string) (userUUID uuid.UUID, deviceUUID uuid.UUID, err error)
}

// NewRoomCreateHandler
// @Summary Создание новой комнаты
// @Description Создаёт новую комнату и возвращает её UUID
// @Tags Room
// @Accept json
// @Produce plain
// @Success 200 {string} string "UUID новой комнаты"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неверный токен"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /room/create [post]
func NewRoomCreateHandler(svc RoomCreator, tokenParser TokenParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr, err := tokenParser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		userUUID, _, err := tokenParser.Parse(tokenStr)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		roomUUID, err := svc.CreateRoom(r.Context(), userUUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(roomUUID.String()))
	}
}

// NewRoomDeleteHandler
// @Summary Удаление комнаты
// @Description Удаляет комнату, если пользователь является владельцем
// @Tags Room
// @Accept json
// @Produce plain
// @Param room-uuid path string true "UUID комнаты"
// @Success 200 "Комната удалена"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неверный токен"
// @Failure 403 "Пользователь не является владельцем"
// @Failure 404 "Комната не найдена"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /room/{room-uuid} [delete]
func NewRoomDeleteHandler(svc RoomDeleter, tokenParser TokenParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomUUIDStr := chi.URLParam(r, "room-uuid")
		roomUUID, err := uuid.Parse(roomUUIDStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenStr, err := tokenParser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		userUUID, _, err := tokenParser.Parse(tokenStr)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		err = svc.DeleteRoom(r.Context(), userUUID, roomUUID)
		if err != nil {
			switch err {
			case services.ErrRoomNotFound:
				w.WriteHeader(http.StatusNotFound)
			case services.ErrNotRoomOwner:
				w.WriteHeader(http.StatusForbidden)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// NewRoomMemberAddHandler
// @Summary Добавление участника в комнату
// @Description Добавляет участника в комнату, если пользователь является владельцем
// @Tags Room
// @Accept json
// @Produce plain
// @Param room-uuid path string true "UUID комнаты"
// @Param member-uuid path string true "UUID участника"
// @Success 200 "Участник добавлен"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неверный токен"
// @Failure 403 "Пользователь не является владельцем"
// @Failure 404 "Комната или участник не найдены"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /room/{room-uuid}/member/{member-uuid}/add [post]
func NewRoomMemberAddHandler(svc RoomMemberAdder, tokenParser TokenParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomUUID, err := uuid.Parse(chi.URLParam(r, "room-uuid"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		memberUUID, err := uuid.Parse(chi.URLParam(r, "member-uuid"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenStr, err := tokenParser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		userUUID, _, err := tokenParser.Parse(tokenStr)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		err = svc.AddMember(r.Context(), userUUID, roomUUID, memberUUID)
		if err != nil {
			switch err {
			case services.ErrRoomNotFound, services.ErrRoomUserNotFound:
				w.WriteHeader(http.StatusNotFound)
			case services.ErrNotRoomOwner:
				w.WriteHeader(http.StatusForbidden)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// NewRoomMemberRemoveHandler
// @Summary Удаление участника из комнаты
// @Description Удаляет участника из комнаты, если пользователь является владельцем
// @Tags Room
// @Accept json
// @Produce plain
// @Param room-uuid path string true "UUID комнаты"
// @Param member-uuid path string true "UUID участника"
// @Success 200 "Участник удалён"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неверный токен"
// @Failure 403 "Пользователь не является владельцем"
// @Failure 404 "Комната или участник не найдены"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /room/{room-uuid}/member/{member-uuid}/remove [post]
func NewRoomMemberRemoveHandler(svc RoomMemberRemover, tokenParser TokenParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomUUID, err := uuid.Parse(chi.URLParam(r, "room-uuid"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		memberUUID, err := uuid.Parse(chi.URLParam(r, "member-uuid"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenStr, err := tokenParser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		userUUID, _, err := tokenParser.Parse(tokenStr)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		err = svc.RemoveMember(r.Context(), userUUID, roomUUID, memberUUID)
		if err != nil {
			switch err {
			case services.ErrRoomNotFound, services.ErrRoomMemberNotFound:
				w.WriteHeader(http.StatusNotFound)
			case services.ErrNotRoomOwner:
				w.WriteHeader(http.StatusForbidden)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // разрешаем подключение с любого origin
}

// NewRoomWebsocketHandler
// @Summary Подключение к комнате через WebSocket
// @Description Устанавливает WebSocket-соединение для отправки сообщений в комнату.
// Пользователь аутентифицируется через JWT токен, передаваемый в заголовке Authorization.
// Сообщения пользователя публикуются в NATS на тему "room.messages".
// @Tags Room
// @Accept json
// @Produce json
// @Param room-uuid path string true "UUID комнаты"
// @Param Authorization header string true "JWT токен пользователя"
// @Success 101 "WebSocket соединение установлено"
// @Failure 400 "Некорректные данные запроса или UUID"
// @Failure 401 "Неверный токен"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /room/{room-uuid}/ws [get]
func NewRoomWebsocketHandler(tokenParser TokenParser, nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomUUID, err := uuid.Parse(chi.URLParam(r, "room-uuid"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tokenStr, err := tokenParser.GetFromRequest(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		userUUID, _, err := tokenParser.Parse(tokenStr)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		for {
			// читаем "сырой" текст из вебсокета
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			chatMsg := models.RoomMessage{
				RoomUUID:  roomUUID,
				UserUUID:  userUUID,
				Message:   string(message),
				Timestamp: time.Now().Unix(),
			}

			data, err := json.Marshal(chatMsg)
			if err != nil {
				continue
			}

			if err := nc.Publish("room.messages", data); err != nil {
				continue
			}
		}
	}
}

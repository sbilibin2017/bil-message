package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
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
		log.Println("[CreateChatHandler] Запрос на создание комнаты получен")

		// Получение токена из запроса
		token, err := parser.GetFromRequest(r)
		if err != nil {
			log.Println("[CreateChatHandler] Ошибка получения токена:", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Println("[CreateChatHandler] Токен получен:", token)

		// Парсинг токена
		userUUID, _, err := parser.Parse(token)
		if err != nil {
			log.Println("[CreateChatHandler] Ошибка парсинга токена:", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Println("[CreateChatHandler] Пользователь успешно аутентифицирован, userUUID:", userUUID)

		// Создание комнаты
		roomUUID, err := svc.CreateRoom(r.Context(), userUUID)
		if err != nil {
			log.Println("[CreateChatHandler] Ошибка создания комнаты:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Println("[CreateChatHandler] Комната успешно создана, roomUUID:", roomUUID)

		// Возврат ответа
		if _, err := w.Write([]byte(roomUUID.String())); err != nil {
			log.Println("[CreateChatHandler] Ошибка отправки ответа:", err)
		}
	}
}

// RemoveChatHandler удаляет комнату по UUID
// @Summary Удаление комнаты
// @Description Удаляет комнату по UUID
// @Tags Chat
// @Accept plain
// @Produce plain
// @Param chat-uuid path string true "UUID комнаты"
// @Success 200 "Комната успешно удалена"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неавторизован"
// @Failure 404 "Комната не найдена"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /chat/{chat-uuid} [delete]
func RemoveChatHandler(svc RoomRemover, parser TokenParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, "chat-uuid")
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

// AddChatMemberHandler добавляет текущего пользователя в комнату
// @Summary Добавление пользователя в комнату
// @Description Добавляет текущего пользователя (из токена) в комнату
// @Tags Chat
// @Accept plain
// @Produce plain
// @Param chat-uuid path string true "UUID комнаты"
// @Success 200 "Пользователь успешно добавлен"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неавторизован"
// @Failure 404 "Комната не найдена"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /chat/{chat-uuid}/member [post]
func AddChatMemberHandler(svc RoomMemberAdder, parser TokenParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, "chat-uuid")
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
		userUUID, _, err := parser.Parse(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := svc.AddRoomMember(r.Context(), roomUUID, userUUID); err != nil {
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

// RemoveChatMemberHandler удаляет текущего пользователя из комнаты
// @Summary Удаление пользователя из комнаты
// @Description Удаляет текущего пользователя (из токена) из комнаты
// @Tags Chat
// @Accept plain
// @Produce plain
// @Param chat-uuid path string true "UUID комнаты"
// @Success 200 "Пользователь успешно удалён"
// @Failure 400 "Некорректные данные запроса"
// @Failure 401 "Неавторизован"
// @Failure 404 "Комната или пользователь не найдены"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /chat/{chat-uuid}/member [delete]
func RemoveChatMemberHandler(svc RoomMemberRemover, parser TokenParser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID := chi.URLParam(r, "chat-uuid")
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
		userUUID, _, err := parser.Parse(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := svc.RemoveRoomMember(r.Context(), roomUUID, userUUID); err != nil {
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

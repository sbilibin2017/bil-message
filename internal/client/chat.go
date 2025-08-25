package client

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// CreateChat отправляет запрос на создание новой комнаты и возвращает UUID комнаты
func CreateChat(ctx context.Context, client *resty.Client, token string) (uuid.UUID, error) {
	token = strings.TrimSpace(token)
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "text/plain").
		SetAuthToken(token).
		Post("/chat")
	if err != nil {
		return uuid.Nil, err
	}

	if resp.IsError() {
		return uuid.Nil, fmt.Errorf("server returned error: %s", resp.Status())
	}

	roomUUID, err := uuid.Parse(resp.String())
	if err != nil {
		return uuid.Nil, err
	}

	return roomUUID, nil
}

// RemoveChat удаляет комнату по UUID
func RemoveChat(ctx context.Context, client *resty.Client, token string, roomUUID uuid.UUID) error {
	token = strings.TrimSpace(token)
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "text/plain").
		SetAuthToken(token).
		Delete("/chat/" + roomUUID.String())
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.Status())
	}

	return nil
}

// AddChatMember добавляет пользователя в указанную комнату
func AddChatMember(ctx context.Context, client *resty.Client, token string, chatUUID uuid.UUID, memberUUID uuid.UUID) error {
	token = strings.TrimSpace(token)
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "text/plain").
		SetAuthToken(token).
		Post("/chat/" + chatUUID.String() + "/" + memberUUID.String())
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.Status())
	}

	return nil
}

// RemoveChatMember удаляет пользователя из указанной комнаты
func RemoveChatMember(ctx context.Context, client *resty.Client, token string, chatUUID uuid.UUID, memberUUID uuid.UUID) error {
	token = strings.TrimSpace(token)
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "text/plain").
		SetAuthToken(token).
		Delete("/chat/" + chatUUID.String() + "/" + memberUUID.String())
	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("server returned error: %s", resp.Status())
	}

	return nil
}

// ConnectWebSocket подключается к указанному wsURL с JWT токеном и запускает чтение/запись сообщений
func ConnectWebSocket(wsURL, token string) error {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+token)

	// подключаемся к WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к WebSocket: %w", err)
	}
	defer conn.Close()

	fmt.Println("WebSocket соединение установлено. Введите сообщения:")

	done := make(chan struct{})

	// Чтение сообщений от сервера
	go func() {
		defer close(done)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Ошибка при чтении:", err)
				return
			}
			fmt.Printf("[Получено] %s\n", string(msg))
		}
	}()

	// Чтение сообщений с консоли и отправка на сервер
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(input)); err != nil {
			fmt.Println("Ошибка отправки:", err)
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Ошибка ввода:", err)
	}

	<-done
	return nil
}

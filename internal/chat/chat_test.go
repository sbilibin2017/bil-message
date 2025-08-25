package chat

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// newTestWSHandler создаёт HTTP-хендлер, который апгрейдит соединение и добавляет клиента в комнату
func newTestWSHandler(room *ChatRoom) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(nil, err) // nil потому что это внутри handler, можно panic

		client := NewChatClient(conn, uuid.New(), room.RoomUUID)
		room.AddClient(client)

		// запускаем горутины внутри методов
		client.ReadPump(room)
		client.WritePump()
	}
}

func TestChatRoomSingleClient(t *testing.T) {
	room := NewChatRoom(uuid.New())
	server := httptest.NewServer(http.HandlerFunc(newTestWSHandler(room)))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):] // http:// -> ws://
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	room.mu.Lock()
	require.Len(t, room.Members, 1)
	room.mu.Unlock()
}

func TestChatRoomMultipleClients(t *testing.T) {
	room := NewChatRoom(uuid.New())
	server := httptest.NewServer(http.HandlerFunc(newTestWSHandler(room)))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]

	c1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer c1.Close()

	c2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer c2.Close()

	// Отправляем сообщение от первого клиента
	testMsg := []byte("hello")
	require.NoError(t, c1.WriteMessage(websocket.TextMessage, testMsg))

	// Проверяем, что второй клиент получил сообщение
	_, msg, err := c2.ReadMessage()
	require.NoError(t, err)
	require.Equal(t, testMsg, msg)

	room.mu.Lock()
	require.Len(t, room.Members, 2)
	room.mu.Unlock()
}

func TestChatClientClose(t *testing.T) {
	room := NewChatRoom(uuid.New())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(nil, err)

		client := NewChatClient(conn, uuid.New(), room.RoomUUID)
		room.AddClient(client)

		client.ReadPump(room)
		client.WritePump()
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	var client *ChatClient
	room.mu.Lock()
	for _, c := range room.Members {
		client = c
		break
	}
	room.mu.Unlock()
	require.NotNil(t, client)

	room.RemoveClient(client)

	room.mu.Lock()
	_, exists := room.Members[client.UserUUID]
	room.mu.Unlock()
	require.False(t, exists)

	// Повторное закрытие безопасно
	client.Close()
	client.Close()
}

func TestChatRoomBroadcastMultipleClients(t *testing.T) {
	room := NewChatRoom(uuid.New())
	server := httptest.NewServer(http.HandlerFunc(newTestWSHandler(room)))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]

	clients := make([]*websocket.Conn, 3)
	for i := 0; i < 3; i++ {
		c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer c.Close()
		clients[i] = c
	}

	testMsg := []byte("broadcast")
	require.NoError(t, clients[0].WriteMessage(websocket.TextMessage, testMsg))

	for i := 1; i < 3; i++ {
		_, msg, err := clients[i].ReadMessage()
		require.NoError(t, err)
		require.Equal(t, testMsg, msg)
	}
}

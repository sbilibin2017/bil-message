package chat

import (
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ChatClient представляет подключение пользователя через WebSocket
type ChatClient struct {
	Conn      *websocket.Conn
	UserUUID  uuid.UUID
	RoomUUID  uuid.UUID
	Send      chan []byte
	closeOnce sync.Once
	closed    chan struct{}
}

// NewChatClient создаёт нового клиента
func NewChatClient(conn *websocket.Conn, userUUID, roomUUID uuid.UUID) *ChatClient {
	return &ChatClient{
		Conn:     conn,
		UserUUID: userUUID,
		RoomUUID: roomUUID,
		Send:     make(chan []byte, 1024),
		closed:   make(chan struct{}),
	}
}

// ReadPump запускает чтение сообщений от клиента в отдельной горутине
func (c *ChatClient) ReadPump(room *ChatRoom) {
	go func() {
		defer c.Close()
		for {
			_, msg, err := c.Conn.ReadMessage()
			if err != nil {
				break
			}
			if room != nil {
				room.Broadcast(msg, c.UserUUID)
			}
		}
	}()
}

// WritePump запускает запись сообщений клиенту в отдельной горутине
func (c *ChatClient) WritePump() {
	go func() {
		for {
			select {
			case msg, ok := <-c.Send:
				if !ok {
					return
				}
				_ = c.Conn.WriteMessage(websocket.TextMessage, msg)
			case <-c.closed:
				return
			}
		}
	}()
}

// Close безопасно закрывает WebSocket соединение и канал Send
func (c *ChatClient) Close() {
	c.closeOnce.Do(func() {
		if c.Conn != nil {
			c.Conn.Close()
		}
		close(c.closed)
		close(c.Send)
	})
}

// ChatRoom представляет комнату с участниками WebSocket
type ChatRoom struct {
	RoomUUID uuid.UUID
	Members  map[uuid.UUID]*ChatClient
	mu       sync.Mutex
}

// NewChatRoom создаёт новую комнату
func NewChatRoom(roomUUID uuid.UUID) *ChatRoom {
	return &ChatRoom{
		RoomUUID: roomUUID,
		Members:  make(map[uuid.UUID]*ChatClient),
	}
}

// AddClient добавляет клиента в комнату
func (r *ChatRoom) AddClient(client *ChatClient) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Members[client.UserUUID] = client
}

// RemoveClient удаляет клиента из комнаты
// RemoveClient удаляет клиента из комнаты
func (r *ChatRoom) RemoveClient(client *ChatClient) {
	r.mu.Lock()
	delete(r.Members, client.UserUUID)
	r.mu.Unlock()
	client.Close()
}

// Broadcast рассылает сообщение всем участникам, кроме отправителя
func (r *ChatRoom) Broadcast(message []byte, senderUUID uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for uid, client := range r.Members {
		if uid != senderUUID {
			select {
			case client.Send <- message:
			default:
				// канал переполнен — сообщение можно пропустить
			}
		}
	}
}

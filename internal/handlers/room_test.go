package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRoomCreateHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockRoomCreator(ctrl)
	mockToken := NewMockTokenParser(ctrl)

	userUUID := uuid.New()
	roomUUID := uuid.New()

	mockToken.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
	mockToken.EXPECT().Parse("token").Return(userUUID, uuid.New(), nil)
	mockSvc.EXPECT().CreateRoom(gomock.Any(), userUUID).Return(roomUUID, nil)

	req := httptest.NewRequest(http.MethodPost, "/room/create", nil)
	w := httptest.NewRecorder()

	handler := NewRoomCreateHandler(mockSvc, mockToken)
	handler.ServeHTTP(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestNewRoomDeleteHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockRoomDeleter(ctrl)
	mockToken := NewMockTokenParser(ctrl)

	userUUID := uuid.New()
	roomUUID := uuid.New()

	mockToken.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
	mockToken.EXPECT().Parse("token").Return(userUUID, uuid.New(), nil)
	mockSvc.EXPECT().DeleteRoom(gomock.Any(), userUUID, roomUUID).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/room/"+roomUUID.String(), nil)
	w := httptest.NewRecorder()

	// chi URLParam работает с context, поэтому создаём router
	r := chi.NewRouter()
	r.Delete("/room/{room-uuid}", NewRoomDeleteHandler(mockSvc, mockToken))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewRoomMemberAddHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockRoomMemberAdder(ctrl)
	mockToken := NewMockTokenParser(ctrl)

	userUUID := uuid.New()
	roomUUID := uuid.New()
	memberUUID := uuid.New()

	mockToken.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
	mockToken.EXPECT().Parse("token").Return(userUUID, uuid.New(), nil)
	mockSvc.EXPECT().AddMember(gomock.Any(), userUUID, roomUUID, memberUUID).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/room/"+roomUUID.String()+"/member/"+memberUUID.String()+"/add", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/room/{room-uuid}/member/{member-uuid}/add", NewRoomMemberAddHandler(mockSvc, mockToken))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestNewRoomMemberRemoveHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockRoomMemberRemover(ctrl)
	mockToken := NewMockTokenParser(ctrl)

	userUUID := uuid.New()
	roomUUID := uuid.New()
	memberUUID := uuid.New()

	mockToken.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
	mockToken.EXPECT().Parse("token").Return(userUUID, uuid.New(), nil)
	mockSvc.EXPECT().RemoveMember(gomock.Any(), userUUID, roomUUID, memberUUID).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/room/"+roomUUID.String()+"/member/"+memberUUID.String()+"/remove", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/room/{room-uuid}/member/{member-uuid}/remove", NewRoomMemberRemoveHandler(mockSvc, mockToken))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// fakeTokenParser для тестирования
type fakeTokenParser struct {
	UserID uuid.UUID
}

func (f *fakeTokenParser) GetFromRequest(r *http.Request) (string, error) {
	return "token", nil
}

func (f *fakeTokenParser) Parse(token string) (uuid.UUID, uuid.UUID, error) {
	return f.UserID, uuid.New(), nil
}

func TestNewWebsocketHandler(t *testing.T) {
	// Поднимаем встроенный NATS сервер
	opts := natsserver.DefaultTestOptions
	opts.Port = -1 // случайный свободный порт
	srv := natsserver.RunServer(&opts)
	defer srv.Shutdown()

	nc, err := nats.Connect(srv.ClientURL())
	require.NoError(t, err)
	defer nc.Close()

	// Канал для проверки сообщений
	msgCh := make(chan models.RoomMessage, 1)
	_, err = nc.Subscribe("room.messages", func(m *nats.Msg) {
		var msg models.RoomMessage
		_ = json.Unmarshal(m.Data, &msg)
		msgCh <- msg
	})
	require.NoError(t, err)

	// fake token parser
	userID := uuid.New()
	parser := &fakeTokenParser{UserID: userID}

	// UUID комнаты
	roomID := uuid.New()

	// Создаём chi Router, чтобы path-параметр "room-uuid" корректно парсился
	r := chi.NewRouter()
	r.Get("/room/{room-uuid}/ws", NewRoomWebsocketHandler(parser, nc))

	// httptest сервер
	server := httptest.NewServer(r)
	defer server.Close()

	// Меняем схему URL для WebSocket
	wsURL := "ws" + server.URL[len("http"):] + "/room/" + roomID.String() + "/ws"

	// Подключаемся
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{
		"Authorization": []string{"Bearer token"},
	})
	require.NoError(t, err)
	defer ws.Close()

	// Отправляем сообщение
	testText := "hello websocket"
	err = ws.WriteJSON(map[string]string{"message": testText})
	require.NoError(t, err)

	// Проверяем, что сообщение пришло через NATS
	select {
	case msg := <-msgCh:
		assert.Equal(t, roomID, msg.RoomUUID)
		assert.Equal(t, userID, msg.UserUUID)
		assert.Equal(t, testText, msg.Message)
		assert.WithinDuration(t, time.Now(), time.Unix(msg.Timestamp, 0), time.Second)
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for message from NATS")
	}
}

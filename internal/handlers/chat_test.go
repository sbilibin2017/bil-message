package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sbilibin2017/bil-message/internal/chat"
	"github.com/sbilibin2017/bil-message/internal/jwt"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateChatHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockRoomCreator(ctrl)
	mockParser := NewMockTokenParser(ctrl)

	userUUID := uuid.New()
	roomUUID := uuid.New()

	tests := []struct {
		name           string
		expectedStatus int
		setup          func()
	}{
		{
			name:           "success",
			expectedStatus: http.StatusOK,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(userUUID, uuid.Nil, nil)
				mockSvc.EXPECT().CreateRoom(gomock.Any(), userUUID).Return(roomUUID, nil)
			},
		},
		{
			name:           "token get error",
			expectedStatus: http.StatusUnauthorized,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("", errors.New("fail"))
			},
		},
		{
			name:           "token parse error",
			expectedStatus: http.StatusUnauthorized,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.Nil, uuid.Nil, errors.New("fail"))
			},
		},
		{
			name:           "service create error",
			expectedStatus: http.StatusInternalServerError,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(userUUID, uuid.Nil, nil)
				mockSvc.EXPECT().CreateRoom(gomock.Any(), userUUID).Return(uuid.Nil, errors.New("fail"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			req := httptest.NewRequest("POST", "/chat", nil)
			w := httptest.NewRecorder()

			handler := CreateChatHandler(mockSvc, mockParser)
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
		})
	}
}

func TestRemoveChatHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockRoomRemover(ctrl)
	mockParser := NewMockTokenParser(ctrl)

	roomUUID := uuid.New()

	tests := []struct {
		name           string
		roomID         string
		expectedStatus int
		setup          func()
	}{
		{
			name:           "success",
			roomID:         roomUUID.String(),
			expectedStatus: http.StatusOK,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.New(), uuid.New(), nil)
				mockSvc.EXPECT().RemoveRoom(gomock.Any(), roomUUID).Return(nil)
			},
		},
		{
			name:           "invalid UUID",
			roomID:         "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
			setup:          func() {},
		},
		{
			name:           "token get error",
			roomID:         roomUUID.String(),
			expectedStatus: http.StatusUnauthorized,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("", errors.New("fail"))
			},
		},
		{
			name:           "token parse error",
			roomID:         roomUUID.String(),
			expectedStatus: http.StatusUnauthorized,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.Nil, uuid.Nil, errors.New("fail"))
			},
		},
		{
			name:           "room not found",
			roomID:         roomUUID.String(),
			expectedStatus: http.StatusNotFound,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.New(), uuid.New(), nil)
				mockSvc.EXPECT().RemoveRoom(gomock.Any(), roomUUID).Return(services.ErrRoomNotFound)
			},
		},
		{
			name:           "internal error",
			roomID:         roomUUID.String(),
			expectedStatus: http.StatusInternalServerError,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.New(), uuid.New(), nil)
				mockSvc.EXPECT().RemoveRoom(gomock.Any(), roomUUID).Return(errors.New("fail"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			r := chi.NewRouter()
			r.Delete("/chat/{room-uuid}", RemoveChatHandler(mockSvc, mockParser))

			req := httptest.NewRequest("DELETE", "/chat/"+tt.roomID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
		})
	}
}

func TestAddChatMemberHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockRoomMemberAdder(ctrl)
	mockParser := NewMockTokenParser(ctrl)

	userUUID := uuid.New()
	roomUUID := uuid.New()

	tests := []struct {
		name           string
		roomID         string
		memberID       string
		expectedStatus int
		setup          func()
	}{
		{
			name:           "success",
			roomID:         roomUUID.String(),
			memberID:       userUUID.String(),
			expectedStatus: http.StatusOK,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.New(), uuid.New(), nil)
				mockSvc.EXPECT().AddRoomMember(gomock.Any(), roomUUID, userUUID).Return(nil)
			},
		},
		{
			name:           "invalid room UUID",
			roomID:         "invalid",
			memberID:       userUUID.String(),
			expectedStatus: http.StatusBadRequest,
			setup:          func() {},
		},
		{
			name:           "invalid member UUID",
			roomID:         roomUUID.String(),
			memberID:       "invalid",
			expectedStatus: http.StatusBadRequest,
			setup:          func() {},
		},
		{
			name:           "token get error",
			roomID:         roomUUID.String(),
			memberID:       userUUID.String(),
			expectedStatus: http.StatusUnauthorized,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("", errors.New("fail"))
			},
		},
		{
			name:           "token parse error",
			roomID:         roomUUID.String(),
			memberID:       userUUID.String(),
			expectedStatus: http.StatusUnauthorized,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.Nil, uuid.Nil, errors.New("fail"))
			},
		},
		{
			name:           "room not found",
			roomID:         roomUUID.String(),
			memberID:       userUUID.String(),
			expectedStatus: http.StatusNotFound,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.New(), uuid.New(), nil)
				mockSvc.EXPECT().AddRoomMember(gomock.Any(), roomUUID, userUUID).Return(services.ErrRoomNotFound)
			},
		},
		{
			name:           "internal error",
			roomID:         roomUUID.String(),
			memberID:       userUUID.String(),
			expectedStatus: http.StatusInternalServerError,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.New(), uuid.New(), nil)
				mockSvc.EXPECT().AddRoomMember(gomock.Any(), roomUUID, userUUID).Return(errors.New("fail"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			r := chi.NewRouter()
			r.Post("/chat/{room-uuid}/{member-uuid}", AddChatMemberHandler(mockSvc, mockParser))

			req := httptest.NewRequest("POST", "/chat/"+tt.roomID+"/"+tt.memberID, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
		})
	}
}

func TestRemoveChatMemberHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := NewMockRoomMemberRemover(ctrl)
	mockParser := NewMockTokenParser(ctrl)

	roomUUID := uuid.New()
	userUUID := uuid.New()

	tests := []struct {
		name           string
		roomID         string
		memberID       string
		expectedStatus int
		setup          func()
	}{
		{
			name:           "success",
			roomID:         roomUUID.String(),
			memberID:       userUUID.String(),
			expectedStatus: http.StatusOK,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.New(), uuid.New(), nil)
				mockSvc.EXPECT().RemoveRoomMember(gomock.Any(), roomUUID, userUUID).Return(nil)
			},
		},
		{
			name:           "invalid room UUID",
			roomID:         "invalid",
			memberID:       userUUID.String(),
			expectedStatus: http.StatusBadRequest,
			setup:          func() {},
		},
		{
			name:           "invalid member UUID",
			roomID:         roomUUID.String(),
			memberID:       "invalid",
			expectedStatus: http.StatusBadRequest,
			setup:          func() {},
		},
		{
			name:           "token get error",
			roomID:         roomUUID.String(),
			memberID:       userUUID.String(),
			expectedStatus: http.StatusUnauthorized,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("", errors.New("fail"))
			},
		},
		{
			name:           "token parse error",
			roomID:         roomUUID.String(),
			memberID:       userUUID.String(),
			expectedStatus: http.StatusUnauthorized,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.Nil, uuid.Nil, errors.New("fail"))
			},
		},
		{
			name:           "room not found",
			roomID:         roomUUID.String(),
			memberID:       userUUID.String(),
			expectedStatus: http.StatusNotFound,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.New(), uuid.New(), nil)
				mockSvc.EXPECT().RemoveRoomMember(gomock.Any(), roomUUID, userUUID).Return(services.ErrRoomNotFound)
			},
		},
		{
			name:           "internal error",
			roomID:         roomUUID.String(),
			memberID:       userUUID.String(),
			expectedStatus: http.StatusInternalServerError,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(uuid.New(), uuid.New(), nil)
				mockSvc.EXPECT().RemoveRoomMember(gomock.Any(), roomUUID, userUUID).Return(errors.New("fail"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			r := chi.NewRouter()
			r.Delete("/chat/{room-uuid}/{member-uuid}", RemoveChatMemberHandler(mockSvc, mockParser))

			req := httptest.NewRequest("DELETE", "/chat/"+tt.roomID+"/"+tt.memberID, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
		})
	}
}

func TestChatWebSocketHandlerWithJWT(t *testing.T) {
	// JWT
	j, err := jwt.New()
	require.NoError(t, err)

	userUUID := uuid.New()
	deviceUUID := uuid.New()
	token, err := j.Generate(userUUID, deviceUUID)
	require.NoError(t, err)

	roomUUID := uuid.New()

	// chi router
	r := chi.NewRouter()
	r.Get("/chat/ws/{room-uuid}", ChatWebSocketHandler(
		func(conn *websocket.Conn, userUUID, roomUUID uuid.UUID) *chat.ChatClient {
			return chat.NewChatClient(conn, userUUID, roomUUID)
		},
		func(roomUUID uuid.UUID) *chat.ChatRoom {
			return chat.NewChatRoom(roomUUID)
		},
		j,
	))

	server := httptest.NewServer(r)
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + server.URL[len("http"):] + "/chat/ws/" + roomUUID.String()

	header := make(map[string][]string)
	header["Authorization"] = []string{"Bearer " + token}

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	require.NoError(t, err)
	defer conn.Close()

	// Send and receive message
	testMsg := []byte("hello")
	require.NoError(t, conn.WriteMessage(websocket.TextMessage, testMsg))

	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	_, _, err = conn.ReadMessage()
	require.Error(t, err) // no other clients yet, timeout expected
}

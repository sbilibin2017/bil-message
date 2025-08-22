package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

// -------------------- NewCreateChatHandler --------------------

func TestNewCreateChatHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJWT := NewMockJWTParser(ctrl)
	mockCreator := NewMockChatCreator(ctrl)

	handler := NewCreateChatHandler(mockJWT, mockCreator)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/chats", nil)
		w := httptest.NewRecorder()

		mockJWT.EXPECT().GetFromRequest(req).Return("token", nil)
		mockJWT.EXPECT().GetUserUUID("token").Return("user-uuid", nil)
		mockCreator.EXPECT().CreateChat(req.Context(), "user-uuid").Return("chat-uuid", nil)

		handler(w, req)

		resp := w.Result()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
	})

	t.Run("unauthorized when token missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/chats", nil)
		w := httptest.NewRecorder()

		mockJWT.EXPECT().GetFromRequest(req).Return("", assert.AnError)

		handler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("forbidden when token invalid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/chats", nil)
		w := httptest.NewRecorder()

		mockJWT.EXPECT().GetFromRequest(req).Return("token", nil)
		mockJWT.EXPECT().GetUserUUID("token").Return("", assert.AnError)

		handler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("internal error when CreateChat fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/chats", nil)
		w := httptest.NewRecorder()

		mockJWT.EXPECT().GetFromRequest(req).Return("token", nil)
		mockJWT.EXPECT().GetUserUUID("token").Return("user-uuid", nil)
		mockCreator.EXPECT().CreateChat(req.Context(), "user-uuid").Return("", assert.AnError)

		handler(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// -------------------- NewAddMemberHandler --------------------

func TestNewAddMemberHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJWT := NewMockJWTParser(ctrl)
	mockAdder := NewMockChatMemberAdder(ctrl)

	handler := NewAddMemberHandler(mockJWT, mockAdder)

	t.Run("success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/chats/123/members", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("chat-uuid", "chat-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		mockJWT.EXPECT().GetFromRequest(req).Return("token", nil)
		mockJWT.EXPECT().GetUserUUID("token").Return("user-uuid", nil)
		mockAdder.EXPECT().AddMember(req.Context(), "chat-uuid", "user-uuid").Return(nil)

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("bad request when chatUUID missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/chats//members", nil)
		w := httptest.NewRecorder()

		mockJWT.EXPECT().GetFromRequest(req).Return("token", nil)
		mockJWT.EXPECT().GetUserUUID("token").Return("user-uuid", nil)

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/chats/123/members", nil)
		w := httptest.NewRecorder()

		mockJWT.EXPECT().GetFromRequest(req).Return("", assert.AnError)

		handler(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/chats/123/members", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("chat-uuid", "chat-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		mockJWT.EXPECT().GetFromRequest(req).Return("token", nil)
		mockJWT.EXPECT().GetUserUUID("token").Return("", assert.AnError)

		handler(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("internal error AddMember", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/chats/123/members", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("chat-uuid", "chat-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		mockJWT.EXPECT().GetFromRequest(req).Return("token", nil)
		mockJWT.EXPECT().GetUserUUID("token").Return("user-uuid", nil)
		mockAdder.EXPECT().AddMember(req.Context(), "chat-uuid", "user-uuid").Return(assert.AnError)

		handler(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestChatWSHandler_Success_MessageBroadcast(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwt := NewMockJWTParser(ctrl)
	reader := NewMockChatReader(ctrl)

	// expectations
	jwt.EXPECT().GetFromRequest(gomock.Any()).Return("valid-token", nil)
	jwt.EXPECT().GetUserUUID("valid-token").Return("user1", nil)
	reader.EXPECT().IsMember(gomock.Any(), "chat123", "user1").Return(true, nil)

	r := chi.NewRouter()
	r.Get("/ws/{chat-uuid}", NewChatWSHandler(jwt, reader))

	srv := httptest.NewServer(r)
	defer srv.Close()

	u := "ws" + srv.URL[len("http"):] + "/ws/chat123"
	ws, resp, err := websocket.DefaultDialer.Dial(u, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	defer ws.Close()

	// send message (no other clients connected)
	err = ws.WriteJSON(map[string]string{"message": "hello"})
	assert.NoError(t, err)
}

func TestChatWSHandler_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwt := NewMockJWTParser(ctrl)
	reader := NewMockChatReader(ctrl)

	// expectation: fail on GetFromRequest
	jwt.EXPECT().GetFromRequest(gomock.Any()).Return("", assert.AnError)

	r := chi.NewRouter()
	r.Get("/ws/{chat-uuid}", NewChatWSHandler(jwt, reader))

	srv := httptest.NewServer(r)
	defer srv.Close()

	u := "ws" + srv.URL[len("http"):] + "/ws/chat123"
	_, resp, err := websocket.DefaultDialer.Dial(u, nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestChatWSHandler_ForbiddenIfNotMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jwt := NewMockJWTParser(ctrl)
	reader := NewMockChatReader(ctrl)

	// expectations
	jwt.EXPECT().GetFromRequest(gomock.Any()).Return("valid-token", nil)
	jwt.EXPECT().GetUserUUID("valid-token").Return("user1", nil)
	reader.EXPECT().IsMember(gomock.Any(), "chat123", "user1").Return(false, nil)

	r := chi.NewRouter()
	r.Get("/ws/{chat-uuid}", NewChatWSHandler(jwt, reader))

	srv := httptest.NewServer(r)
	defer srv.Close()

	u := "ws" + srv.URL[len("http"):] + "/ws/chat123"
	_, resp, err := websocket.DefaultDialer.Dial(u, nil)
	assert.Error(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func newTestServerWithParam(t *testing.T, jwt *MockJWTParser, reader *MockChatReader, chatUUID string) *httptest.Server {
	r := chi.NewRouter()
	r.Get("/ws/{chat-uuid}", func(w http.ResponseWriter, r *http.Request) {
		// Подменяем chat-uuid
		rc := chi.NewRouteContext()
		rc.URLParams.Add("chat-uuid", chatUUID)
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rc))
		NewChatWSHandler(jwt, reader)(w, r)
	})
	return httptest.NewServer(r)
}

func TestChatWSHandler_ErrorCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		setupMocks     func(jwt *MockJWTParser, reader *MockChatReader)
		chatUUID       string
		expectedStatus int
	}{
		{
			name: "Unauthorized when GetFromRequest fails",
			setupMocks: func(jwt *MockJWTParser, reader *MockChatReader) {
				jwt.EXPECT().GetFromRequest(gomock.Any()).Return("", errors.New("bad token"))
			},
			chatUUID:       "chat123",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Forbidden when GetUserUUID fails",
			setupMocks: func(jwt *MockJWTParser, reader *MockChatReader) {
				jwt.EXPECT().GetFromRequest(gomock.Any()).Return("valid", nil)
				jwt.EXPECT().GetUserUUID("valid").Return("", errors.New("parse error"))
			},
			chatUUID:       "chat123",
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "BadRequest when chatUUID missing",
			setupMocks: func(jwt *MockJWTParser, reader *MockChatReader) {
				jwt.EXPECT().GetFromRequest(gomock.Any()).Return("valid", nil)
				jwt.EXPECT().GetUserUUID("valid").Return("user1", nil)
			},
			chatUUID:       "", // пустой UUID → BadRequest
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "InternalServerError when IsMember returns error",
			setupMocks: func(jwt *MockJWTParser, reader *MockChatReader) {
				jwt.EXPECT().GetFromRequest(gomock.Any()).Return("valid", nil)
				jwt.EXPECT().GetUserUUID("valid").Return("user1", nil)
				reader.EXPECT().IsMember(gomock.Any(), "chat123", "user1").Return(false, errors.New("db error"))
			},
			chatUUID:       "chat123",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Forbidden when user is not a member",
			setupMocks: func(jwt *MockJWTParser, reader *MockChatReader) {
				jwt.EXPECT().GetFromRequest(gomock.Any()).Return("valid", nil)
				jwt.EXPECT().GetUserUUID("valid").Return("user1", nil)
				reader.EXPECT().IsMember(gomock.Any(), "chat123", "user1").Return(false, nil)
			},
			chatUUID:       "chat123",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwt := NewMockJWTParser(ctrl)
			reader := NewMockChatReader(ctrl)

			tt.setupMocks(jwt, reader)

			srv := newTestServerWithParam(t, jwt, reader, tt.chatUUID)
			defer srv.Close()

			resp, err := http.Get(srv.URL + "/ws/dummy")
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

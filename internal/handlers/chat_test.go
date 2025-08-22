package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/websocket"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateChatHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJWT := NewMockJWTParser(ctrl)
	mockSvc := NewMockChatCreator(ctrl)
	handler := NewCreateChatHandler(mockJWT, mockSvc)

	validToken := "token-123"
	userUUID := "user-uuid-1"
	chatUUID := "chat-uuid-1"

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			mockSetup: func() {
				mockJWT.EXPECT().GetFromRequest(gomock.Any()).Return(&validToken, nil)
				mockJWT.EXPECT().Parse(validToken).Return(&models.TokenPayload{UserUUID: userUUID}, nil)
				mockSvc.EXPECT().CreateChat(gomock.Any(), userUUID).Return(&chatUUID, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   chatUUID,
		},
		{
			name: "unauthorized - no token",
			mockSetup: func() {
				mockJWT.EXPECT().GetFromRequest(gomock.Any()).Return(nil, errors.New("no token"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "forbidden - invalid token",
			mockSetup: func() {
				mockJWT.EXPECT().GetFromRequest(gomock.Any()).Return(&validToken, nil)
				mockJWT.EXPECT().Parse(validToken).Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "internal error - CreateChat failed",
			mockSetup: func() {
				mockJWT.EXPECT().GetFromRequest(gomock.Any()).Return(&validToken, nil)
				mockJWT.EXPECT().Parse(validToken).Return(&models.TokenPayload{UserUUID: userUUID}, nil)
				mockSvc.EXPECT().CreateChat(gomock.Any(), userUUID).Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockSetup != nil {
				tt.mockSetup()
			}

			req := httptest.NewRequest(http.MethodPost, "/chats/create", nil)
			w := httptest.NewRecorder()

			handler(w, req)
			resp := w.Result()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.expectedBody != "" {
				buf := new(bytes.Buffer)
				buf.ReadFrom(resp.Body)
				assert.Equal(t, tt.expectedBody, buf.String())
			}
		})
	}
}

func TestAddMemberHandler(t *testing.T) {
	validToken := "token-123"
	chatUUID := "chat-uuid-1"
	userUUID := "user-uuid-2"

	tests := []struct {
		name           string
		url            string
		mockSetup      func(mockJWT *MockJWTParser, mockSvc *MockChatMemberAdder)
		expectedStatus int
	}{
		{
			name: "/ success",
			url:  "/chats/" + chatUUID + "/members/" + userUUID,
			mockSetup: func(mockJWT *MockJWTParser, mockSvc *MockChatMemberAdder) {
				mockJWT.EXPECT().GetFromRequest(gomock.Any()).Return(&validToken, nil)
				mockJWT.EXPECT().Parse(validToken).Return(&models.TokenPayload{UserUUID: "creator-uuid"}, nil)
				mockSvc.EXPECT().AddMember(gomock.Any(), chatUUID, userUUID).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "/ unauthorized - no token",
			url:  "/chats/" + chatUUID + "/members/" + userUUID,
			mockSetup: func(mockJWT *MockJWTParser, mockSvc *MockChatMemberAdder) {
				mockJWT.EXPECT().GetFromRequest(gomock.Any()).Return(nil, errors.New("no token"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "/ forbidden - invalid token",
			url:  "/chats/" + chatUUID + "/members/" + userUUID,
			mockSetup: func(mockJWT *MockJWTParser, mockSvc *MockChatMemberAdder) {
				mockJWT.EXPECT().GetFromRequest(gomock.Any()).Return(&validToken, nil)
				mockJWT.EXPECT().Parse(validToken).Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "/ bad request - missing params",
			url:  "/chats//members/",
			mockSetup: func(mockJWT *MockJWTParser, mockSvc *MockChatMemberAdder) {
				// Настраиваем возвращение токена, чтобы код дошёл до проверки URL-параметров
				token := "token-123"
				mockJWT.EXPECT().GetFromRequest(gomock.Any()).Return(&token, nil)
				mockJWT.EXPECT().Parse(token).Return(&models.TokenPayload{UserUUID: "creator-uuid"}, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "/ internal error - AddMember failed",
			url:  "/chats/" + chatUUID + "/members/" + userUUID,
			mockSetup: func(mockJWT *MockJWTParser, mockSvc *MockChatMemberAdder) {
				mockJWT.EXPECT().GetFromRequest(gomock.Any()).Return(&validToken, nil)
				mockJWT.EXPECT().Parse(validToken).Return(&models.TokenPayload{UserUUID: "creator-uuid"}, nil)
				mockSvc.EXPECT().AddMember(gomock.Any(), chatUUID, userUUID).Return(errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockJWT := NewMockJWTParser(ctrl)
			mockSvc := NewMockChatMemberAdder(ctrl)
			handler := NewAddMemberHandler(mockJWT, mockSvc)

			if tt.mockSetup != nil {
				tt.mockSetup(mockJWT, mockSvc)
			}

			req := httptest.NewRequest(http.MethodPost, tt.url, nil)

			// Если нужны URL-параметры
			if tt.url != "/chats//members/" {
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("chat-uuid", chatUUID)
				rctx.URLParams.Add("user-uuid", userUUID)
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			}

			w := httptest.NewRecorder()
			handler(w, req)
			resp := w.Result()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestChatWSHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJWT := NewMockJWTParser(ctrl)
	userUUID := "user-1"
	chatUUID := "chat-1"
	validToken := "token-123"

	handler := NewChatWSHandler(mockJWT)

	// Chi router for path params
	r := chi.NewRouter()
	r.Get("/chats/{chat-uuid}/ws", handler)

	server := httptest.NewServer(r)
	defer server.Close()

	wsURL := "ws" + server.URL[4:] + "/chats/" + chatUUID + "/ws"

	// Set up mocks
	mockJWT.EXPECT().GetFromRequest(gomock.Any()).Return(&validToken, nil)
	mockJWT.EXPECT().Parse(validToken).Return(&models.TokenPayload{UserUUID: userUUID}, nil)

	// Connect
	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if assert.NoError(t, err) {
		defer conn.Close()
		assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
	}
}

package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/stretchr/testify/assert"
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
		expectedStatus int
		setup          func()
	}{
		{
			name:           "success",
			roomID:         roomUUID.String(),
			expectedStatus: http.StatusOK,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(userUUID, uuid.Nil, nil)
				mockSvc.EXPECT().AddRoomMember(gomock.Any(), roomUUID, userUUID).Return(nil)
			},
		},
		{
			name:           "invalid UUID",
			roomID:         "invalid",
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
				mockParser.EXPECT().Parse("token").Return(userUUID, uuid.Nil, nil)
				mockSvc.EXPECT().AddRoomMember(gomock.Any(), roomUUID, userUUID).Return(services.ErrRoomNotFound)
			},
		},
		{
			name:           "add user error",
			roomID:         roomUUID.String(),
			expectedStatus: http.StatusInternalServerError,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(userUUID, uuid.Nil, nil)
				mockSvc.EXPECT().AddRoomMember(gomock.Any(), roomUUID, userUUID).Return(errors.New("fail"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			r := chi.NewRouter()
			r.Post("/chat/{chat-uuid}/member", AddChatMemberHandler(mockSvc, mockParser))

			req := httptest.NewRequest("POST", "/chat/"+tt.roomID+"/member", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

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
			r.Delete("/chat/{chat-uuid}", RemoveChatHandler(mockSvc, mockParser))

			req := httptest.NewRequest("DELETE", "/chat/"+tt.roomID, nil)
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
		expectedStatus int
		setup          func()
	}{
		{
			name:           "success",
			roomID:         roomUUID.String(),
			expectedStatus: http.StatusOK,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(userUUID, uuid.New(), nil)
				mockSvc.EXPECT().RemoveRoomMember(gomock.Any(), roomUUID, userUUID).Return(nil)
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
				mockParser.EXPECT().Parse("token").Return(userUUID, uuid.New(), nil)
				mockSvc.EXPECT().RemoveRoomMember(gomock.Any(), roomUUID, userUUID).Return(services.ErrRoomNotFound)
			},
		},
		{
			name:           "internal error",
			roomID:         roomUUID.String(),
			expectedStatus: http.StatusInternalServerError,
			setup: func() {
				mockParser.EXPECT().GetFromRequest(gomock.Any()).Return("token", nil)
				mockParser.EXPECT().Parse("token").Return(userUUID, uuid.New(), nil)
				mockSvc.EXPECT().RemoveRoomMember(gomock.Any(), roomUUID, userUUID).Return(errors.New("fail"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()

			r := chi.NewRouter()
			r.Delete("/chat/{chat-uuid}/member", RemoveChatMemberHandler(mockSvc, mockParser))

			req := httptest.NewRequest("DELETE", "/chat/"+tt.roomID+"/member", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Result().StatusCode)
		})
	}
}

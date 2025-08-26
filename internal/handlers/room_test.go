package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

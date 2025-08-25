package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestRoomService_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRW := NewMockRoomWriter(ctrl)
	mockRR := NewMockRoomReader(ctrl)
	mockRMW := NewMockRoomMemberWriter(ctrl)
	mockRMR := NewMockRoomMemberReader(ctrl)
	svc := NewChatService(mockRW, mockRR, mockRMW, mockRMR)
	userUUID := uuid.New()
	ctx := context.Background()

	tests := []struct {
		name          string
		mockSetup     func()
		expectedError error
	}{
		{
			name: "success",
			mockSetup: func() {
				mockRW.EXPECT().Save(gomock.Any(), gomock.Any(), userUUID).Return(nil)
				mockRMW.EXPECT().Save(gomock.Any(), gomock.Any(), userUUID, gomock.Any()).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "room writer save error",
			mockSetup: func() {
				mockRW.EXPECT().Save(gomock.Any(), gomock.Any(), userUUID).Return(errors.New("room save fail"))
			},
			expectedError: errors.New("room save fail"),
		},
		{
			name: "room member save error",
			mockSetup: func() {
				mockRW.EXPECT().Save(gomock.Any(), gomock.Any(), userUUID).Return(nil)
				mockRMW.EXPECT().Save(gomock.Any(), gomock.Any(), userUUID, gomock.Any()).Return(errors.New("save member fail"))
			},
			expectedError: errors.New("save member fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			roomUUID, err := svc.CreateRoom(ctx, userUUID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Equal(t, uuid.Nil, roomUUID)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, roomUUID)
			}
		})
	}
}

func TestRoomService_AddUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRW := NewMockRoomWriter(ctrl)
	mockRR := NewMockRoomReader(ctrl)
	mockRMW := NewMockRoomMemberWriter(ctrl)
	mockRMR := NewMockRoomMemberReader(ctrl)
	svc := NewChatService(mockRW, mockRR, mockRMW, mockRMR)
	roomUUID := uuid.New()
	userUUID := uuid.New()
	ctx := context.Background()

	tests := []struct {
		name          string
		mockRoom      *models.RoomDB
		mockRoomErr   error
		mockSaveErr   error
		expectedError error
	}{
		{"success", &models.RoomDB{RoomUUID: roomUUID}, nil, nil, nil},
		{"room not found", nil, nil, nil, ErrRoomNotFound},
		{"room reader error", nil, errors.New("db error"), nil, errors.New("db error")},
		{"save error", &models.RoomDB{RoomUUID: roomUUID}, nil, errors.New("save fail"), errors.New("save fail")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRR.EXPECT().Get(gomock.Any(), roomUUID).Return(tt.mockRoom, tt.mockRoomErr)
			if tt.mockRoom != nil {
				mockRMW.EXPECT().Save(gomock.Any(), roomUUID, userUUID, gomock.Any()).Return(tt.mockSaveErr)
			}

			err := svc.AddRoomMember(ctx, roomUUID, userUUID)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRoomService_RemoveUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRW := NewMockRoomWriter(ctrl)
	mockRR := NewMockRoomReader(ctrl)
	mockRMW := NewMockRoomMemberWriter(ctrl)
	mockRMR := NewMockRoomMemberReader(ctrl)
	svc := NewChatService(mockRW, mockRR, mockRMW, mockRMR)
	roomUUID := uuid.New()
	userUUID := uuid.New()
	member := &models.RoomMemberDB{
		RoomUUID: roomUUID,
		UserUUID: userUUID,
		JoinedAt: time.Now(),
	}
	ctx := context.Background()

	tests := []struct {
		name          string
		mockRoom      *models.RoomDB
		mockRoomErr   error
		mockMember    *models.RoomMemberDB
		mockMemberErr error
		mockDelErr    error
		expectedError error
	}{
		{"success", &models.RoomDB{RoomUUID: roomUUID}, nil, member, nil, nil, nil},
		{"room not found", nil, nil, nil, nil, nil, ErrRoomNotFound},
		{"room reader error", nil, errors.New("db error"), nil, nil, nil, errors.New("db error")},
		{"member not in room", &models.RoomDB{RoomUUID: roomUUID}, nil, nil, nil, nil, ErrUserNotInRoom},
		{"member get error", &models.RoomDB{RoomUUID: roomUUID}, nil, nil, errors.New("member read fail"), nil, errors.New("member read fail")},
		{"delete error", &models.RoomDB{RoomUUID: roomUUID}, nil, member, nil, errors.New("delete fail"), errors.New("delete fail")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRR.EXPECT().Get(gomock.Any(), roomUUID).Return(tt.mockRoom, tt.mockRoomErr)
			if tt.mockRoom != nil {
				mockRMR.EXPECT().Get(gomock.Any(), roomUUID, userUUID).Return(tt.mockMember, tt.mockMemberErr)
			}
			if tt.mockMember != nil {
				mockRMW.EXPECT().Delete(gomock.Any(), roomUUID, userUUID).Return(tt.mockDelErr)
			}

			err := svc.RemoveRoomMember(ctx, roomUUID, userUUID)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRoomService_Remove(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRW := NewMockRoomWriter(ctrl)
	mockRR := NewMockRoomReader(ctrl)
	mockRMW := NewMockRoomMemberWriter(ctrl)
	mockRMR := NewMockRoomMemberReader(ctrl)
	svc := NewChatService(mockRW, mockRR, mockRMW, mockRMR)
	roomUUID := uuid.New()
	ctx := context.Background()

	tests := []struct {
		name          string
		mockRoom      *models.RoomDB
		mockRoomErr   error
		mockDelErr    error
		expectedError error
	}{
		{
			name:          "success",
			mockRoom:      &models.RoomDB{RoomUUID: roomUUID},
			mockRoomErr:   nil,
			mockDelErr:    nil,
			expectedError: nil,
		},
		{
			name:          "room not found",
			mockRoom:      nil,
			mockRoomErr:   nil,
			mockDelErr:    nil,
			expectedError: ErrRoomNotFound,
		},
		{
			name:          "room reader error",
			mockRoom:      nil,
			mockRoomErr:   errors.New("db error"),
			mockDelErr:    nil,
			expectedError: errors.New("db error"),
		},
		{
			name:          "delete error",
			mockRoom:      &models.RoomDB{RoomUUID: roomUUID},
			mockRoomErr:   nil,
			mockDelErr:    errors.New("delete fail"),
			expectedError: errors.New("delete fail"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRR.EXPECT().Get(gomock.Any(), roomUUID).Return(tt.mockRoom, tt.mockRoomErr)
			if tt.mockRoom != nil {
				mockRW.EXPECT().Delete(gomock.Any(), roomUUID).Return(tt.mockDelErr)
			}

			err := svc.RemoveRoom(ctx, roomUUID)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

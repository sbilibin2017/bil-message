package services

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestChatService_CreateChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockChatWriter(ctrl)
	mockReader := NewMockChatReader(ctrl)

	svc := NewChatService(mockWriter, mockReader)

	creatorUUID := "user-123"
	// Save должен вызваться с chatUUID, creatorUUID, creatorUUID
	mockWriter.EXPECT().
		Save(gomock.Any(), gomock.Any(), creatorUUID, creatorUUID).
		Return(nil)

	chatUUID, err := svc.CreateChat(context.Background(), creatorUUID)

	assert.NoError(t, err)
	assert.NotEmpty(t, chatUUID)
}

func TestChatService_AddMember_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockChatWriter(ctrl)
	mockReader := NewMockChatReader(ctrl)

	svc := NewChatService(mockWriter, mockReader)

	chatUUID := "chat-1"
	creatorUUID := "user-123"
	newMemberUUID := "user-456"

	chat := &models.ChatDB{
		ChatUUID:          chatUUID,
		CreatedByUUID:     creatorUUID,
		ParticipantsUUIDs: creatorUUID,
	}

	// Сначала Get вернёт чат
	mockReader.EXPECT().Get(gomock.Any(), chatUUID).Return(chat, nil)

	// Потом Save вызовется с обновлёнными участниками
	mockWriter.EXPECT().
		Save(gomock.Any(), chatUUID, creatorUUID, creatorUUID+","+newMemberUUID).
		Return(nil)

	err := svc.AddMember(context.Background(), chatUUID, newMemberUUID)

	assert.NoError(t, err)
}

func TestChatService_AddMember_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockChatWriter(ctrl)
	mockReader := NewMockChatReader(ctrl)

	svc := NewChatService(mockWriter, mockReader)

	chatUUID := "chat-1"
	creatorUUID := "user-123"

	chat := &models.ChatDB{
		ChatUUID:          chatUUID,
		CreatedByUUID:     creatorUUID,
		ParticipantsUUIDs: creatorUUID,
	}

	mockReader.EXPECT().Get(gomock.Any(), chatUUID).Return(chat, nil)

	// Save не должен вызываться
	err := svc.AddMember(context.Background(), chatUUID, creatorUUID)

	assert.NoError(t, err)
}

func TestChatService_AddMember_ChatNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockChatWriter(ctrl)
	mockReader := NewMockChatReader(ctrl)

	svc := NewChatService(mockWriter, mockReader)

	chatUUID := "chat-1"

	mockReader.EXPECT().Get(gomock.Any(), chatUUID).Return(nil, nil)

	err := svc.AddMember(context.Background(), chatUUID, "user-456")

	assert.Error(t, err)
	assert.Equal(t, "chat not found", err.Error())
}

func TestChatService_IsMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockWriter := NewMockChatWriter(ctrl)
	mockReader := NewMockChatReader(ctrl)

	svc := NewChatService(mockWriter, mockReader)

	chatUUID := "chat-1"
	userUUID := "user-123"

	tests := []struct {
		name       string
		chat       *models.ChatDB
		mockErr    error
		wantMember bool
		wantErr    bool
	}{
		{
			name: "user is member",
			chat: &models.ChatDB{
				ChatUUID:          chatUUID,
				ParticipantsUUIDs: "user-123,user-456",
			},
			wantMember: true,
		},
		{
			name: "user not member",
			chat: &models.ChatDB{
				ChatUUID:          chatUUID,
				ParticipantsUUIDs: "user-999,user-456",
			},
			wantMember: false,
		},
		{
			name:    "chat not found",
			chat:    nil,
			wantErr: true,
		},
		{
			name:    "reader error",
			mockErr: errors.New("db error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockReader.EXPECT().Get(gomock.Any(), chatUUID).Return(tt.chat, tt.mockErr)

			isMember, err := svc.IsMember(context.Background(), chatUUID, userUUID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantMember, isMember)
			}
		})
	}
}

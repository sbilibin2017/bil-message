package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestChatService_CreateChat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockChatWriter := services.NewMockChatWriter(ctrl)
	mockMemberWriter := services.NewMockChatMemberWriter(ctrl)

	svc := services.NewChatService(mockChatWriter, mockMemberWriter)
	ctx := context.Background()
	userUUID := "user-123"

	t.Run("successful chat creation", func(t *testing.T) {
		mockChatWriter.EXPECT().
			Save(ctx, gomock.Any()).
			Return(nil)
		mockMemberWriter.EXPECT().
			Save(ctx, gomock.Any()).
			Return(nil)

		chatUUID, err := svc.CreateChat(ctx, userUUID)
		assert.NoError(t, err)
		assert.NotNil(t, chatUUID)
	})

	t.Run("chat save error", func(t *testing.T) {
		mockChatWriter.EXPECT().
			Save(ctx, gomock.Any()).
			Return(errors.New("chat save error"))

		chatUUID, err := svc.CreateChat(ctx, userUUID)
		assert.Error(t, err)
		assert.Nil(t, chatUUID)
	})

	t.Run("member save error", func(t *testing.T) {
		mockChatWriter.EXPECT().
			Save(ctx, gomock.Any()).
			Return(nil)
		mockMemberWriter.EXPECT().
			Save(ctx, gomock.Any()).
			Return(errors.New("member save error"))

		chatUUID, err := svc.CreateChat(ctx, userUUID)
		assert.Error(t, err)
		assert.Nil(t, chatUUID)
	})
}

func TestChatService_AddMember(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMemberWriter := services.NewMockChatMemberWriter(ctrl)
	svc := services.NewChatService(nil, mockMemberWriter)

	ctx := context.Background()
	chatUUID := "chat-123"
	userUUID := "user-456"

	t.Run("successful add member", func(t *testing.T) {
		mockMemberWriter.EXPECT().
			Save(ctx, gomock.Any()).
			Return(nil)

		err := svc.AddMember(ctx, chatUUID, userUUID)
		assert.NoError(t, err)
	})

	t.Run("add member save error", func(t *testing.T) {
		mockMemberWriter.EXPECT().
			Save(ctx, gomock.Any()).
			Return(errors.New("member save error"))

		err := svc.AddMember(ctx, chatUUID, userUUID)
		assert.Error(t, err)
		assert.EqualError(t, err, "member save error")
	})
}

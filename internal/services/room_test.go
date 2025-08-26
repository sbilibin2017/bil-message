package services_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
	"github.com/sbilibin2017/bil-message/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestRoomService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	roomUUID := uuid.New()
	ownerUUID := uuid.New()
	memberUUID := uuid.New()
	otherUserUUID := uuid.New()

	mockRoomWriter := services.NewMockRoomWriter(ctrl)
	mockRoomReader := services.NewMockRoomReader(ctrl)
	mockMemberWriter := services.NewMockRoomMemberWriter(ctrl)
	mockMemberReader := services.NewMockRoomMemberReader(ctrl)
	mockUserReader := services.NewMockRoomUserReader(ctrl)

	service := services.NewRoomService(
		mockRoomWriter,
		mockRoomReader,
		mockMemberWriter,
		mockMemberReader,
		mockUserReader,
	)

	t.Run("CreateRoom", func(t *testing.T) {
		mockRoomWriter.EXPECT().Save(ctx, gomock.Any(), ownerUUID).Return(nil)

		roomID, err := service.CreateRoom(ctx, ownerUUID)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, roomID)
	})

	t.Run("DeleteRoom_OwnerCheck", func(t *testing.T) {
		t.Run("owner can delete", func(t *testing.T) {
			mockRoomReader.EXPECT().
				Get(ctx, roomUUID).
				Return(&models.RoomDB{RoomUUID: roomUUID, OwnerUUID: ownerUUID}, nil)

			mockRoomWriter.EXPECT().Delete(ctx, roomUUID).Return(nil)

			err := service.DeleteRoom(ctx, ownerUUID, roomUUID)
			assert.NoError(t, err)
		})

		t.Run("non-owner cannot delete", func(t *testing.T) {
			mockRoomReader.EXPECT().
				Get(ctx, roomUUID).
				Return(&models.RoomDB{RoomUUID: roomUUID, OwnerUUID: ownerUUID}, nil)

			err := service.DeleteRoom(ctx, otherUserUUID, roomUUID)
			assert.ErrorIs(t, err, services.ErrNotRoomOwner)
		})
	})

	t.Run("AddMember_OwnerCheck", func(t *testing.T) {
		t.Run("owner can add member", func(t *testing.T) {
			mockUserReader.EXPECT().GetByUUID(ctx, memberUUID).
				Return(&models.UserDB{UserUUID: memberUUID}, nil)

			mockRoomReader.EXPECT().Get(ctx, roomUUID).
				Return(&models.RoomDB{RoomUUID: roomUUID, OwnerUUID: ownerUUID}, nil)

			mockMemberWriter.EXPECT().Save(ctx, roomUUID, memberUUID).Return(nil)

			err := service.AddMember(ctx, ownerUUID, roomUUID, memberUUID)
			assert.NoError(t, err)
		})

		t.Run("non-owner cannot add member", func(t *testing.T) {
			mockUserReader.EXPECT().GetByUUID(ctx, memberUUID).
				Return(&models.UserDB{UserUUID: memberUUID}, nil)

			mockRoomReader.EXPECT().Get(ctx, roomUUID).
				Return(&models.RoomDB{RoomUUID: roomUUID, OwnerUUID: ownerUUID}, nil)

			err := service.AddMember(ctx, otherUserUUID, roomUUID, memberUUID)
			assert.ErrorIs(t, err, services.ErrNotRoomOwner)
		})
	})

	t.Run("RemoveMember_OwnerCheck", func(t *testing.T) {
		t.Run("owner can remove member", func(t *testing.T) {
			mockRoomReader.EXPECT().Get(ctx, roomUUID).
				Return(&models.RoomDB{RoomUUID: roomUUID, OwnerUUID: ownerUUID}, nil)

			mockMemberReader.EXPECT().Get(ctx, roomUUID, memberUUID).
				Return(&models.RoomMemberDB{RoomUUID: roomUUID, MemberUUID: memberUUID}, nil)

			mockMemberWriter.EXPECT().Delete(ctx, roomUUID, memberUUID).Return(nil)

			err := service.RemoveMember(ctx, ownerUUID, roomUUID, memberUUID)
			assert.NoError(t, err)
		})

		t.Run("non-owner cannot remove member", func(t *testing.T) {
			mockRoomReader.EXPECT().Get(ctx, roomUUID).
				Return(&models.RoomDB{RoomUUID: roomUUID, OwnerUUID: ownerUUID}, nil)

			err := service.RemoveMember(ctx, otherUserUUID, roomUUID, memberUUID)
			assert.ErrorIs(t, err, services.ErrNotRoomOwner)
		})
	})
}

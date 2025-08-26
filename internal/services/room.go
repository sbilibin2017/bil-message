package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// Ошибки сервиса комнат
var (
	ErrRoomNotFound       = errors.New("room does not exist")
	ErrRoomMemberNotFound = errors.New("room member does not exist")
	ErrRoomUserNotFound   = errors.New("user does not exist")
	ErrNotRoomOwner       = errors.New("user is not the room owner")
)

// RoomWriter интерфейс для записи комнат в БД
type RoomWriter interface {
	Save(ctx context.Context, roomUUID, ownerUUID uuid.UUID) error
	Delete(ctx context.Context, roomUUID uuid.UUID) error
}

// RoomReader интерфейс для чтения комнат из БД
type RoomReader interface {
	Get(ctx context.Context, roomUUID uuid.UUID) (*models.RoomDB, error)
}

// RoomMemberWriter интерфейс для записи участников комнаты в БД
type RoomMemberWriter interface {
	Save(ctx context.Context, roomUUID, memberUUID uuid.UUID) error
	Delete(ctx context.Context, roomUUID, memberUUID uuid.UUID) error
}

// RoomMemberReader интерфейс для чтения участников комнаты из БД
type RoomMemberReader interface {
	Get(ctx context.Context, roomUUID, memberUUID uuid.UUID) (*models.RoomMemberDB, error)
}

// RoomUserReader интерфейс для чтения пользователей из БД
type RoomUserReader interface {
	GetByUUID(ctx context.Context, userUUID uuid.UUID) (*models.UserDB, error)
}

// RoomService сервис управления комнатами
type RoomService struct {
	roomWriteRepo       RoomWriter
	roomReadRepo        RoomReader
	roomMemberWriteRepo RoomMemberWriter
	roomMemberReadRepo  RoomMemberReader
	userReader          RoomUserReader
}

func NewRoomService(
	roomWriteRepo RoomWriter,
	roomReadRepo RoomReader,
	roomMemberWriteRepo RoomMemberWriter,
	roomMemberReadRepo RoomMemberReader,
	userReader RoomUserReader,
) *RoomService {
	return &RoomService{
		roomWriteRepo:       roomWriteRepo,
		roomReadRepo:        roomReadRepo,
		roomMemberWriteRepo: roomMemberWriteRepo,
		roomMemberReadRepo:  roomMemberReadRepo,
		userReader:          userReader,
	}
}

// CreateRoom создаёт новую комнату с владельцем
func (s *RoomService) CreateRoom(ctx context.Context, userUUID uuid.UUID) (uuid.UUID, error) {
	roomUUID := uuid.New()
	if err := s.roomWriteRepo.Save(ctx, roomUUID, userUUID); err != nil {
		return uuid.Nil, err
	}
	return roomUUID, nil
}

// DeleteRoom удаляет комнату только если userUUID владелец
func (s *RoomService) DeleteRoom(ctx context.Context, userUUID, roomUUID uuid.UUID) error {
	room, err := s.roomReadRepo.Get(ctx, roomUUID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrRoomNotFound
	}
	if room.OwnerUUID != userUUID {
		return ErrNotRoomOwner
	}

	return s.roomWriteRepo.Delete(ctx, roomUUID)
}

// AddMember добавляет участника в комнату только если userUUID владелец
func (s *RoomService) AddMember(ctx context.Context, userUUID, roomUUID, memberUUID uuid.UUID) error {
	user, err := s.userReader.GetByUUID(ctx, memberUUID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrRoomUserNotFound
	}

	room, err := s.roomReadRepo.Get(ctx, roomUUID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrRoomNotFound
	}
	if room.OwnerUUID != userUUID {
		return ErrNotRoomOwner
	}

	return s.roomMemberWriteRepo.Save(ctx, roomUUID, memberUUID)
}

// RemoveMember удаляет участника из комнаты только если userUUID владелец
func (s *RoomService) RemoveMember(ctx context.Context, userUUID, roomUUID, memberUUID uuid.UUID) error {
	room, err := s.roomReadRepo.Get(ctx, roomUUID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrRoomNotFound
	}
	if room.OwnerUUID != userUUID {
		return ErrNotRoomOwner
	}

	member, err := s.roomMemberReadRepo.Get(ctx, roomUUID, memberUUID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrRoomMemberNotFound
	}
	return s.roomMemberWriteRepo.Delete(ctx, roomUUID, memberUUID)
}

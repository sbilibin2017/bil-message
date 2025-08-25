package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// ErrRoomNotFound возвращается, если комната с указанным UUID не найдена.
var ErrRoomNotFound = errors.New("room not found")

// ErrUserNotInRoom возвращается, если пользователь не состоит в комнате.
var ErrUserNotInRoom = errors.New("user not in room")

// RoomWriter описывает интерфейс для создания или обновления комнаты в хранилище.
type RoomWriter interface {
	// Save сохраняет комнату с указанным roomUUID и creatorUUID.
	// Если комната уже существует, обновляет только updated_at.
	Save(ctx context.Context, roomUUID uuid.UUID, creatorUUID uuid.UUID) error

	Delete(ctx context.Context, roomUUID uuid.UUID) error
}

// RoomReader описывает интерфейс для чтения информации о комнате.
type RoomReader interface {
	// Get возвращает комнату по roomUUID или nil, если комната не найдена.
	Get(ctx context.Context, roomUUID uuid.UUID) (*models.RoomDB, error)
}

// RoomMemberWriter описывает интерфейс для добавления и удаления участников комнаты.
type RoomMemberWriter interface {
	// Save добавляет пользователя в комнату.
	// Если запись уже существует, обновляет только updated_at.
	Save(ctx context.Context, roomUUID uuid.UUID, userUUID uuid.UUID, joinedAt time.Time) error

	// Delete удаляет пользователя из комнаты.
	Delete(ctx context.Context, roomUUID uuid.UUID, userUUID uuid.UUID) error
}

// RoomMemberReader описывает интерфейс для чтения участников комнаты.
type RoomMemberReader interface {
	// Get возвращает участника комнаты по roomUUID и userUUID, или nil если не найден.
	Get(ctx context.Context, roomUUID, userUUID uuid.UUID) (*models.RoomMemberDB, error)
}

// ChatService реализует бизнес-логику работы с комнатами и их участниками.
type ChatService struct {
	rw  RoomWriter       // репозиторий для записи комнат
	rr  RoomReader       // репозиторий для чтения комнат
	rmw RoomMemberWriter // репозиторий для записи участников
	rmr RoomMemberReader // репозиторий для чтения участников
}

// NewChatService создаёт новый экземпляр RoomService с указанными репозиториями.
func NewChatService(
	rw RoomWriter,
	rr RoomReader,
	rmw RoomMemberWriter,
	rmr RoomMemberReader,
) *ChatService {
	return &ChatService{
		rw:  rw,
		rr:  rr,
		rmw: rmw,
		rmr: rmr,
	}
}

// Create создаёт новую комнату и добавляет создателя как участника.
func (svc *ChatService) CreateRoom(ctx context.Context, userUUID uuid.UUID) (roomUUID uuid.UUID, err error) {
	roomUUID = uuid.New()

	if err := svc.rw.Save(ctx, roomUUID, userUUID); err != nil {
		return uuid.Nil, err
	}

	if err := svc.rmw.Save(ctx, roomUUID, userUUID, time.Now().UTC()); err != nil {
		return uuid.Nil, err
	}

	return roomUUID, nil
}

// Remove удаляет комнату и всех её участников.
func (svc *ChatService) RemoveRoom(ctx context.Context, roomUUID uuid.UUID) error {
	room, err := svc.rr.Get(ctx, roomUUID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrRoomNotFound
	}

	return svc.rw.Delete(ctx, roomUUID)
}

// AddUser добавляет пользователя в существующую комнату..
func (svc *ChatService) AddRoomMember(ctx context.Context, roomUUID uuid.UUID, userUUID uuid.UUID) error {
	room, err := svc.rr.Get(ctx, roomUUID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrRoomNotFound
	}

	return svc.rmw.Save(ctx, roomUUID, userUUID, time.Now().UTC())
}

// RemoveUser удаляет пользователя из комнаты.
func (svc *ChatService) RemoveRoomMember(ctx context.Context, roomUUID uuid.UUID, userUUID uuid.UUID) error {
	room, err := svc.rr.Get(ctx, roomUUID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrRoomNotFound
	}

	member, err := svc.rmr.Get(ctx, roomUUID, userUUID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrUserNotInRoom
	}

	return svc.rmw.Delete(ctx, roomUUID, userUUID)
}

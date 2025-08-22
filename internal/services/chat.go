package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// ChatWriter отвечает за работу с таблицей chats
type ChatWriter interface {
	Save(
		ctx context.Context,
		chatUUID string,
		createdByUUID string,
		participantsUUIDs string,
	) error
}

type ChatReader interface {
	Get(
		ctx context.Context,
		chatUUID string,
	) (*models.ChatDB, error)
}

// ChatService управляет бизнес-логикой чатов
type ChatService struct {
	cw ChatWriter
	cr ChatReader
}

// NewChatService создаёт ChatService
func NewChatService(
	cw ChatWriter,
	cr ChatReader,
) *ChatService {
	return &ChatService{
		cw: cw,
		cr: cr,
	}
}

// CreateChat создаёт новый чат и добавляет создателя в качестве участника
func (s *ChatService) CreateChat(
	ctx context.Context,
	createdByUUID string,
) (chatUUID string, err error) {
	chatUUID = uuid.New().String()
	if err := s.cw.Save(ctx, chatUUID, createdByUUID, createdByUUID); err != nil {
		return "", err
	}
	return chatUUID, nil
}

// AddMember добавляет нового участника в чат
func (s *ChatService) AddMember(
	ctx context.Context,
	chatUUID string,
	participantUUID string,
) error {
	chat, err := s.cr.Get(ctx, chatUUID)
	if err != nil {
		return err
	}
	if chat == nil {
		return errors.New("chat not found")
	}

	participantsSlice := strings.Split(chat.ParticipantsUUIDs, ",")
	for _, p := range participantsSlice {
		if p == participantUUID {
			return nil
		}
	}

	participantsSlice = append(participantsSlice, participantUUID)
	updatedParticipants := strings.Join(participantsSlice, ",")

	return s.cw.Save(ctx, chatUUID, chat.CreatedByUUID, updatedParticipants)
}

// IsMember проверяет, является ли пользователь участником чата
func (s *ChatService) IsMember(
	ctx context.Context,
	chatUUID string,
	userUUID string,
) (bool, error) {
	chat, err := s.cr.Get(ctx, chatUUID)
	if err != nil {
		return false, err
	}
	if chat == nil {
		return false, errors.New("chat not found")
	}

	participantsSlice := strings.Split(chat.ParticipantsUUIDs, ",")
	for _, p := range participantsSlice {
		if p == userUUID {
			return true, nil
		}
	}
	return false, nil
}

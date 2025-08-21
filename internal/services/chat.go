package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// ChatWriter отвечает за работу с таблицей chats
type ChatWriter interface {
	Save(ctx context.Context, chat *models.ChatDB) error
}

// ChatMemberWriter отвечает за работу с таблицей chat_members
type ChatMemberWriter interface {
	Save(ctx context.Context, chatMember *models.ChatMemberDB) error
}

// ChatService управляет бизнес-логикой чатов
type ChatService struct {
	chats   ChatWriter
	members ChatMemberWriter
}

// NewChatService создаёт ChatService
func NewChatService(chats ChatWriter, members ChatMemberWriter) *ChatService {
	return &ChatService{
		chats:   chats,
		members: members,
	}
}

// CreateChat создаёт новый чат и добавляет создателя в качестве участника
func (s *ChatService) CreateChat(ctx context.Context, userUUID string) (*string, error) {
	chatUUID := uuid.New().String()

	// Создаём объект ChatDB
	chat := &models.ChatDB{
		ChatUUID:      chatUUID,
		CreatedByUUID: userUUID,
	}

	// Сохраняем чат
	if err := s.chats.Save(ctx, chat); err != nil {
		return nil, err
	}

	// Создаём объект ChatMemberDB для создателя
	member := &models.ChatMemberDB{
		ChatMemberUUID: uuid.New().String(),
		ChatUUID:       chatUUID,
		UserUUID:       userUUID,
	}

	// Добавляем участника
	if err := s.members.Save(ctx, member); err != nil {
		return nil, err
	}

	return &chatUUID, nil
}

// AddMember добавляет нового участника в чат
func (s *ChatService) AddMember(ctx context.Context, chatUUID string, userUUID string) error {
	// Создаём объект ChatMemberDB для создателя
	member := &models.ChatMemberDB{
		ChatMemberUUID: uuid.New().String(),
		ChatUUID:       chatUUID,
		UserUUID:       userUUID,
	}
	return s.members.Save(ctx, member)
}

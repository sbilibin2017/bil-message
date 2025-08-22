package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// ChatMemberWriteRepository реализует интерфейс chat.ChatMemberWriter через sqlx.DB
type ChatMemberWriteRepository struct {
	db *sqlx.DB
}

// NewChatMemberWriteRepository создаёт новый ChatMemberWriteRepository
func NewChatMemberWriteRepository(db *sqlx.DB) *ChatMemberWriteRepository {
	return &ChatMemberWriteRepository{db: db}
}

// Save сохраняет члена чата в таблицу chat_members или обновляет при конфликте
func (r *ChatMemberWriteRepository) Save(
	ctx context.Context,
	chatMember *models.ChatMemberDB,
) error {
	query := `
		INSERT INTO chat_members (chat_member_uuid, chat_uuid, user_uuid, joined_at, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(chat_uuid, user_uuid) DO UPDATE
		SET updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query, chatMember.ChatMemberUUID, chatMember.ChatUUID, chatMember.UserUUID)
	return err
}

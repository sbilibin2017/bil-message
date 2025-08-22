package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/bil-message/internal/models"
)

// ChatWriteRepository реализует интерфейс ChatWriter через sqlx.DB
type ChatWriteRepository struct {
	db *sqlx.DB
}

// NewChatWriteRepository создаёт новый ChatWriteRepository
func NewChatWriteRepository(db *sqlx.DB) *ChatWriteRepository {
	return &ChatWriteRepository{db: db}
}

// Save сохраняет чат в таблицу chats
func (r *ChatWriteRepository) Save(
	ctx context.Context,
	chat *models.ChatDB,
) error {
	query := `
		INSERT INTO chats (chat_uuid, created_by_uuid, created_at, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(chat_uuid) DO UPDATE
		SET updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query, chat.ChatUUID, chat.CreatedByUUID)
	return err
}

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

func NewChatWriteRepository(db *sqlx.DB) *ChatWriteRepository {
	return &ChatWriteRepository{db: db}
}

// Save сохраняет чат в таблицу chats
func (r *ChatWriteRepository) Save(
	ctx context.Context,
	chatUUID string,
	createdByUUID string,
	participantsUUIDs string,
) error {
	query := `
		INSERT INTO chats (chat_uuid, participants_uuids, created_by_uuid, created_at, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(chat_uuid) DO UPDATE
		SET 			
			participants_uuids = excluded.participants_uuids,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(
		ctx,
		query,
		chatUUID,
		participantsUUIDs,
		createdByUUID,
	)
	return err
}

// ChatReadRepository реализует интерфейс ChatReader через sqlx.DB
type ChatReadRepository struct {
	db *sqlx.DB
}

func NewChatReadRepository(db *sqlx.DB) *ChatReadRepository {
	return &ChatReadRepository{db: db}
}

// Get возвращает чат по его chat_uuid
func (r *ChatReadRepository) Get(ctx context.Context, chatUUID string) (*models.ChatDB, error) {
	var chat models.ChatDB
	query := `
		SELECT chat_uuid, participants_uuids, created_by_uuid, created_at, updated_at
		FROM chats
		WHERE chat_uuid = ?
	`
	err := r.db.GetContext(ctx, &chat, query, chatUUID)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

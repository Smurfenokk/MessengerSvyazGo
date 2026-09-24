package repository

import (
	"context"
	"database/sql"

	"messenger-svyaz/internal/db/sqlc"
)

type MessageRepository struct {
	q *sqlc.Queries
}

func NewMessageRepository(q *sqlc.Queries) *MessageRepository {
	return &MessageRepository{q: q}
}

func (r *MessageRepository) Create(ctx context.Context, id, sender, receiver, content, msgType string, replyTo *string) error {
	var rt sql.NullString
	if replyTo != nil {
		rt = sql.NullString{String: *replyTo, Valid: true}
	}

	return r.q.CreateMessage(ctx, sqlc.CreateMessageParams{
		ID:       id,
		Sender:   sender,
		Receiver: receiver,
		Content:  content,
		Type:     msgType,
		ReplyTo:  rt,
	})
}

func (r *MessageRepository) Get(ctx context.Context, id string) (*sqlc.Message, error) {
	return r.q.GetMessage(ctx, id)
}

func (r *MessageRepository) GetChatMessages(ctx context.Context, user1, user2 string) ([]sqlc.Message, error) {
	return r.q.GetChatMessages(ctx, sqlc.GetChatMessagesParams{
		Column1: user1,
		Column2: user2,
	})
}

func (r *MessageRepository) GetSavedMessages(ctx context.Context, username string) ([]sqlc.Message, error) {
	return r.q.GetSavedMessages(ctx, username)
}

func (r *MessageRepository) GetRecentChats(ctx context.Context, username string) ([]sqlc.GetRecentChatsRow, error) {
	return r.q.GetRecentChats(ctx, username)
}

func (r *MessageRepository) MarkAsRead(ctx context.Context, id string) error {
	return r.q.UpdateMessageRead(ctx, id)
}

func (r *MessageRepository) UpdateContent(ctx context.Context, id, content string) error {
	return r.q.UpdateMessageEdited(ctx, sqlc.UpdateMessageEditedParams{
		Content: content,
		Column2: id,
	})
}

func (r *MessageRepository) MarkAsDeleted(ctx context.Context, id string) error {
	return r.q.UpdateMessageDeleted(ctx, id)
}

func (r *MessageRepository) UpdateReactions(ctx context.Context, id string, reactions interface{}) error {
	return r.q.UpdateMessageReaction(ctx, sqlc.UpdateMessageReactionParams{
		Column2: id,
		Reactions: reactions, // TODO: handle JSONB properly
	})
}

func (r *MessageRepository) GetUnreadCount(ctx context.Context, username string) (int64, error) {
	return r.q.GetUnreadCount(ctx, username)
}

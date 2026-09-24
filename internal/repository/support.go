package repository

import (
	"context"

	"messenger-svyaz/internal/db/sqlc"
)

type SupportRepository struct {
	q *sqlc.Queries
}

func NewSupportRepository(q *sqlc.Queries) *SupportRepository {
	return &SupportRepository{q: q}
}

func (r *SupportRepository) Create(ctx context.Context, id, user, sender, content, msgType string) error {
	return r.q.CreateSupportMessage(ctx, sqlc.CreateSupportMessageParams{
		ID:     id,
		User:   user,
		Sender: sender,
		Content: content,
		Type:   msgType,
	})
}

func (r *SupportRepository) GetByUser(ctx context.Context, username string) ([]sqlc.SupportMessage, error) {
	return r.q.GetSupportMessagesByUser(ctx, username)
}

func (r *SupportRepository) GetAll(ctx context.Context) ([]sqlc.SupportMessage, error) {
	return r.q.GetAllSupportMessages(ctx)
}

func (r *SupportRepository) MarkAsAnswered(ctx context.Context, id string) error {
	return r.q.UpdateSupportMessageAnswered(ctx, id)
}

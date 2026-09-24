package repository

import (
	"context"
	"database/sql"
	"crypto/rand"
	"math/big"
	"time"

	"messenger-svyaz/internal/db/sqlc"
)

type KeyRepository struct {
	q *sqlc.Queries
}

func NewKeyRepository(q *sqlc.Queries) *KeyRepository {
	return &KeyRepository{q: q}
}

func (r *KeyRepository) Create(ctx context.Context, key string) error {
	return r.q.CreateRegistrationKey(ctx, key)
}

func (r *KeyRepository) Get(ctx context.Context, key string) (*sqlc.RegistrationKey, error) {
	return r.q.GetRegistrationKey(ctx, key)
}

func (r *KeyRepository) Use(ctx context.Context, username, key string) error {
	return r.q.UseRegistrationKey(ctx, sqlc.UseRegistrationKeyParams{
		UsedBy: username,
		Column2: key,
	})
}

func (r *KeyRepository) List(ctx context.Context) ([]sqlc.RegistrationKey, error) {
	return r.q.ListRegistrationKeys(ctx)
}

func (r *KeyRepository) Deactivate(ctx context.Context, key string) error {
	return r.q.DeactivateRegistrationKey(ctx, key)
}

func (r *KeyRepository) GenerateAndCreate(ctx context.Context, count int) ([]string, error) {
	keys := make([]string, 0, count)
	for i := 0; i < count; i++ {
		key := generateKey()
		if err := r.Create(ctx, key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func generateKey() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 16)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Fallback to time-based if crypto/rand fails
			n = big.NewInt(int64(time.Now().UnixNano() % int64(len(charset))))
		}
		b[i] = charset[n.Int64()]
	}
	return string(b)
}

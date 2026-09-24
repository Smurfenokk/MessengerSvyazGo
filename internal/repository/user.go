package repository

import (
	"context"
	"database/sql"

	"messenger-svyaz/internal/db/sqlc"
)

type UserRepository struct {
	q *sqlc.Queries
}

func NewUserRepository(q *sqlc.Queries) *UserRepository {
	return &UserRepository{q: q}
}

func (r *UserRepository) Create(ctx context.Context, username, accountID, passwordHash, firstName, lastName, middleName, displayName string, isAdmin bool, bio string) error {
	var ln, mn sql.NullString
	if lastName != "" {
		ln = sql.NullString{String: lastName, Valid: true}
	}
	if middleName != "" {
		mn = sql.NullString{String: middleName, Valid: true}
	}

	return r.q.CreateUser(ctx, sqlc.CreateUserParams{
		Username:     username,
		AccountID:    accountID,
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     ln,
		MiddleName:   mn,
		DisplayName:  displayName,
		IsAdmin:      isAdmin,
		Bio:          bio,
	})
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*sqlc.User, error) {
	return r.q.GetUserByUsername(ctx, username)
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, username string) error {
	return r.q.UpdateUserLastLogin(ctx, username)
}

func (r *UserRepository) UpdateOnlineStatus(ctx context.Context, username string, isOnline bool) error {
	return r.q.UpdateUserOnlineStatus(ctx, sqlc.UpdateUserOnlineStatusParams{
		Username: username,
		Column2:  isOnline,
	})
}

func (r *UserRepository) UpdateBlocked(ctx context.Context, username string, isBlocked bool) error {
	return r.q.UpdateUserBlocked(ctx, sqlc.UpdateUserBlockedParams{
		Username: username,
		Column2:  isBlocked,
	})
}

func (r *UserRepository) UpdateBio(ctx context.Context, username, bio string) error {
	return r.q.UpdateUserBio(ctx, sqlc.UpdateUserBioParams{
		Bio:      bio,
		Username: username,
	})
}

func (r *UserRepository) UpdatePassword(ctx context.Context, username, passwordHash string) error {
	return r.q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		Passwordhash: passwordHash,
		Username:     username,
	})
}

func (r *UserRepository) UpdateAvatar(ctx context.Context, username, avatarKey string, hasAvatar bool) error {
	var ak sql.NullString
	if avatarKey != "" {
		ak = sql.NullString{String: avatarKey, Valid: true}
	}

	return r.q.UpdateUserAvatar(ctx, sqlc.UpdateUserAvatarParams{
		AvatarKey:  ak,
		Column2:    hasAvatar,
		Username:   username,
	})
}

func (r *UserRepository) List(ctx context.Context, excludeAdmins bool, search string) ([]sqlc.ListUsersRow, error) {
	return r.q.ListUsers(ctx, sqlc.ListUsersParams{
		Column1: excludeAdmins,
		Column2: nullableString(search),
	})
}

func (r *UserRepository) ListAll(ctx context.Context) ([]sqlc.User, error) {
	return r.q.ListAllUsers(ctx)
}

func nullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

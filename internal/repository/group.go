package repository

import (
	"context"
	"database/sql"

	"messenger-svyaz/internal/db/sqlc"
)

type GroupRepository struct {
	q *sqlc.Queries
}

func NewGroupRepository(q *sqlc.Queries) *GroupRepository {
	return &GroupRepository{q: q}
}

func (r *GroupRepository) Create(ctx context.Context, id, name, description, creator string, settings interface{}) error {
	var desc sql.NullString
	if description != "" {
		desc = sql.NullString{String: description, Valid: true}
	}

	return r.q.CreateGroup(ctx, sqlc.CreateGroupParams{
		ID:          id,
		Name:        name,
		Description: desc,
		Creator:     creator,
		Settings:    settings, // TODO: handle JSONB properly
	})
}

func (r *GroupRepository) Get(ctx context.Context, id string) (*sqlc.Group, error) {
	return r.q.GetGroup(ctx, id)
}

func (r *GroupRepository) GetByCreator(ctx context.Context, creator string) ([]sqlc.Group, error) {
	return r.q.GetGroupsByCreator(ctx, creator)
}

func (r *GroupRepository) AddMember(ctx context.Context, groupID, username string) error {
	return r.q.AddGroupMember(ctx, sqlc.AddGroupMemberParams{
		GroupID:  groupID,
		Username: username,
	})
}

func (r *GroupRepository) RemoveMember(ctx context.Context, groupID, username string) error {
	return r.q.RemoveGroupMember(ctx, sqlc.RemoveGroupMemberParams{
		GroupID:  groupID,
		Username: username,
	})
}

func (r *GroupRepository) GetMembers(ctx context.Context, groupID string) ([]string, error) {
	return r.q.GetGroupMembers(ctx, groupID)
}

func (r *GroupRepository) GetWithMember(ctx context.Context, username string) ([]sqlc.GetGroupsWithMemberRow, error) {
	return r.q.GetGroupsWithMember(ctx, username)
}

func (r *GroupRepository) UpdateSettings(ctx context.Context, id string, settings interface{}) error {
	return r.q.UpdateGroupSettings(ctx, sqlc.UpdateGroupSettingsParams{
		Settings: settings, // TODO: handle JSONB properly
		Column2:  id,
	})
}

func (r *GroupRepository) CreateMessage(ctx context.Context, id, groupID, sender, content, msgType string, replyTo *string) error {
	var rt sql.NullString
	if replyTo != nil {
		rt = sql.NullString{String: *replyTo, Valid: true}
	}

	return r.q.CreateGroupMessage(ctx, sqlc.CreateGroupMessageParams{
		ID:       id,
		GroupID:  groupID,
		Sender:   sender,
		Content:  content,
		Type:     msgType,
		ReplyTo:  rt,
	})
}

func (r *GroupRepository) GetMessage(ctx context.Context, id string) (*sqlc.GroupMessage, error) {
	return r.q.GetGroupMessage(ctx, id)
}

func (r *GroupRepository) GetMessages(ctx context.Context, groupID string) ([]sqlc.GroupMessage, error) {
	return r.q.GetGroupMessages(ctx, groupID)
}

func (r *GroupRepository) UpdateMessageReaction(ctx context.Context, id string, reactions interface{}) error {
	return r.q.UpdateGroupMessageReaction(ctx, sqlc.UpdateGroupMessageReactionParams{
		Column2: id,
		Reactions: reactions, // TODO: handle JSONB properly
	})
}

func (r *GroupRepository) UpdateMessageContent(ctx context.Context, id, content string) error {
	return r.q.UpdateGroupMessageEdited(ctx, sqlc.UpdateGroupMessageEditedParams{
		Content: content,
		Column2: id,
	})
}

func (r *GroupRepository) MarkMessageAsDeleted(ctx context.Context, id string) error {
	return r.q.UpdateGroupMessageDeleted(ctx, id)
}

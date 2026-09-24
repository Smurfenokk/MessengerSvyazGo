-- name: CreateGroup :exec
INSERT INTO groups (id, name, description, creator, settings)
VALUES ($1, $2, $3, $4, $5);

-- name: GetGroup :one
SELECT * FROM groups WHERE id = $1;

-- name: GetGroupsByCreator :many
SELECT * FROM groups WHERE creator = $1 ORDER BY created_at DESC;

-- name: AddGroupMember :exec
INSERT INTO group_members (group_id, username) VALUES ($1, $2);

-- name: RemoveGroupMember :exec
DELETE FROM group_members WHERE group_id = $1 AND username = $2;

-- name: GetGroupMembers :many
SELECT username FROM group_members WHERE group_id = $1;

-- name: GetGroupsWithMember :many
SELECT g.*, gm.username as member_username
FROM groups g
JOIN group_members gm ON g.id = gm.group_id
WHERE gm.username = $1;

-- name: UpdateGroupSettings :exec
UPDATE groups SET settings = $1 WHERE id = $2;

-- name: CreateGroupMessage :exec
INSERT INTO group_messages (id, group_id, sender, content, type, reply_to)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetGroupMessage :one
SELECT * FROM group_messages WHERE id = $1;

-- name: GetGroupMessages :many
SELECT * FROM group_messages WHERE group_id = $1 ORDER BY timestamp ASC;

-- name: UpdateGroupMessageReaction :exec
UPDATE group_messages SET reactions = $1 WHERE id = $2;

-- name: UpdateGroupMessageEdited :exec
UPDATE group_messages SET edited = true, content = $1 WHERE id = $2;

-- name: UpdateGroupMessageDeleted :exec
UPDATE group_messages SET deleted = true WHERE id = $1;

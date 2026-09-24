-- name: CreateMessage :exec
INSERT INTO messages (id, sender, receiver, content, type, reply_to)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetMessage :one
SELECT * FROM messages WHERE id = $1;

-- name: GetChatMessages :many
SELECT * FROM messages
WHERE (sender = $1 AND receiver = $2)
   OR (sender = $2 AND receiver = $1)
ORDER BY timestamp ASC;

-- name: GetSavedMessages :many
SELECT * FROM messages
WHERE sender = $1 AND type = 'saved'
ORDER BY timestamp ASC;

-- name: GetRecentChats :many
SELECT DISTINCT
    CASE
        WHEN sender = $1 THEN receiver
        ELSE sender
    END as other_user,
    m.*
FROM messages m
WHERE (sender = $1 OR receiver = $1) AND type != 'saved'
ORDER BY timestamp DESC;

-- name: UpdateMessageRead :exec
UPDATE messages SET read = true WHERE id = $1;

-- name: UpdateMessageEdited :exec
UPDATE messages SET edited = true, content = $1 WHERE id = $2;

-- name: UpdateMessageDeleted :exec
UPDATE messages SET deleted = true WHERE id = $1;

-- name: UpdateMessageReaction :exec
UPDATE messages SET reactions = $1 WHERE id = $2;

-- name: GetUnreadCount :one
SELECT COUNT(*) FROM messages
WHERE receiver = $1 AND read = false AND deleted = false;

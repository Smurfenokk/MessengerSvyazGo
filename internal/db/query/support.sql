-- name: CreateSupportMessage :exec
INSERT INTO support_messages (id, user, sender, content, type)
VALUES ($1, $2, $3, $4, $5);

-- name: GetSupportMessagesByUser :many
SELECT * FROM support_messages WHERE user = $1 ORDER BY timestamp ASC;

-- name: GetAllSupportMessages :many
SELECT * FROM support_messages ORDER BY timestamp DESC;

-- name: UpdateSupportMessageAnswered :exec
UPDATE support_messages SET answered = true WHERE id = $1;

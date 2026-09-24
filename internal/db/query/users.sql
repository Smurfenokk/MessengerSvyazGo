-- name: CreateUser :exec
INSERT INTO users (username, account_id, password_hash, first_name, last_name, middle_name, display_name, is_admin, bio)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: UpdateUserLastLogin :exec
UPDATE users SET last_login = NOW() WHERE username = $1;

-- name: UpdateUserOnlineStatus :exec
UPDATE users SET is_online = $2 WHERE username = $1;

-- name: UpdateUserBlocked :exec
UPDATE users SET is_blocked = $2 WHERE username = $1;

-- name: UpdateUserBio :exec
UPDATE users SET bio = $1 WHERE username = $2;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $1 WHERE username = $2;

-- name: UpdateUserAvatar :exec
UPDATE users SET avatar_key = $1, has_avatar = $2 WHERE username = $3;

-- name: ListUsers :many
SELECT username, account_id, display_name, is_admin, is_online, is_blocked, bio, avatar_key
FROM users
WHERE ($1::bool IS FALSE OR is_admin = false)
  AND ($2::text IS NULL OR username = $2)
ORDER BY username;

-- name: ListAllUsers :many
SELECT * FROM users ORDER BY created_at DESC;

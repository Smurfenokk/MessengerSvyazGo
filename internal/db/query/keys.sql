-- name: CreateRegistrationKey :exec
INSERT INTO registration_keys (key, created_at, is_active)
VALUES ($1, NOW(), true);

-- name: GetRegistrationKey :one
SELECT * FROM registration_keys WHERE key = $1;

-- name: UseRegistrationKey :exec
UPDATE registration_keys
SET used_by = $1, used_at = NOW(), is_active = false
WHERE key = $2;

-- name: ListRegistrationKeys :many
SELECT * FROM registration_keys ORDER BY created_at DESC;

-- name: DeactivateRegistrationKey :exec
UPDATE registration_keys SET is_active = false WHERE key = $1;

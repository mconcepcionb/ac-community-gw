-- name: InsertAPIKey :one
INSERT INTO api_keys (id, name, key_prefix, key_hash, permissions)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListAPIKeys :many
SELECT * FROM api_keys ORDER BY created_at DESC;

-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1;

-- name: TouchAPIKey :exec
UPDATE api_keys SET last_used_at = now() WHERE id = $1;

-- name: RotateAPIKey :one
UPDATE api_keys SET key_prefix = $2, key_hash = $3, permissions = $4
WHERE id = $1
RETURNING *;

-- name: DeleteAPIKey :exec
DELETE FROM api_keys WHERE id = $1;

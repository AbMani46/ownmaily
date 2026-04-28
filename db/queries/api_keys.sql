-- name: CreateAPIKey :one
INSERT INTO api_keys (key_hash, key_prefix)
VALUES ($1, $2)
RETURNING *;

-- name: GetLatestAPIKey :one
SELECT * FROM api_keys ORDER BY created_at DESC LIMIT 1;

-- name: DeleteAllAPIKeys :exec
DELETE FROM api_keys;

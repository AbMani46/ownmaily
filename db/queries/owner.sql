-- name: CreateOwner :one
INSERT INTO owner (email, password_hash)
VALUES ($1, $2)
RETURNING *;

-- name: GetOwner :one
SELECT * FROM owner WHERE id = TRUE;

-- name: UpdateOwnerPassword :exec
UPDATE owner SET password_hash = $1, updated_at = NOW() WHERE id = TRUE;

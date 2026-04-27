-- name: AddSuppression :exec
INSERT INTO suppressed_emails (email, reason)
VALUES ($1, $2)
ON CONFLICT (email) DO NOTHING;

-- name: GetSuppression :one
SELECT * FROM suppressed_emails WHERE email = $1;

-- name: ListSuppressions :many
SELECT * FROM suppressed_emails
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountSuppressions :one
SELECT COUNT(*) FROM suppressed_emails;

-- name: DeleteSuppression :exec
DELETE FROM suppressed_emails WHERE email = $1;

-- name: IsSuppressed :one
SELECT EXISTS(SELECT 1 FROM suppressed_emails WHERE email = $1);

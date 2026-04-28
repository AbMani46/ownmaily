-- name: CreateSubscriber :one
INSERT INTO subscribers (email, first_name, last_name, status, source)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSubscriberByID :one
SELECT * FROM subscribers WHERE id = $1;

-- name: GetSubscriberByEmail :one
SELECT * FROM subscribers WHERE email = $1;

-- name: ListSubscribers :many
SELECT * FROM subscribers
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountSubscribers :one
SELECT COUNT(*) FROM subscribers;

-- name: UpdateSubscriber :one
UPDATE subscribers SET
    email      = $2,
    first_name = $3,
    last_name  = $4,
    status     = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateSubscriberStatus :exec
UPDATE subscribers SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: DeleteSubscriber :exec
DELETE FROM subscribers WHERE id = $1;

-- name: SearchSubscribers :many
SELECT * FROM subscribers
WHERE email ILIKE $1 OR first_name ILIKE $1 OR last_name ILIKE $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListSubscribersByStatus :many
SELECT * FROM subscribers
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllSubscribers :many
SELECT * FROM subscribers
ORDER BY created_at DESC;

-- name: CountSubscribersByStatus :one
SELECT COUNT(*) FROM subscribers WHERE status = $1;

-- name: UpsertSubscriber :one
INSERT INTO subscribers (email, first_name, last_name, status, source)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (email) DO UPDATE SET
    first_name = EXCLUDED.first_name,
    last_name  = EXCLUDED.last_name,
    updated_at = NOW()
RETURNING *;

-- name: CreateList :one
INSERT INTO lists (name, description, double_opt_in)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetListByID :one
SELECT * FROM lists WHERE id = $1;

-- name: ListLists :many
SELECT * FROM lists ORDER BY name ASC;

-- name: UpdateList :one
UPDATE lists SET
    name          = $2,
    description   = $3,
    double_opt_in = $4,
    updated_at    = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteList :exec
DELETE FROM lists WHERE id = $1;

-- name: CountSubscribersInList :one
SELECT COUNT(*) FROM list_subscribers WHERE list_id = $1;

-- name: AddSubscriberToList :exec
INSERT INTO list_subscribers (list_id, subscriber_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveSubscriberFromList :exec
DELETE FROM list_subscribers WHERE list_id = $1 AND subscriber_id = $2;

-- name: ListSubscribersInList :many
SELECT s.* FROM subscribers s
JOIN list_subscribers ls ON ls.subscriber_id = s.id
WHERE ls.list_id = $1
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

-- name: IsSubscriberInList :one
SELECT EXISTS(
    SELECT 1 FROM list_subscribers WHERE list_id = $1 AND subscriber_id = $2
);

-- name: CreateTag :one
INSERT INTO tags (name) VALUES ($1) RETURNING *;

-- name: GetTagByID :one
SELECT * FROM tags WHERE id = $1;

-- name: GetTagByName :one
SELECT * FROM tags WHERE name = $1;

-- name: ListTags :many
SELECT * FROM tags ORDER BY name ASC;

-- name: UpdateTag :one
UPDATE tags SET name = $2 WHERE id = $1 RETURNING *;

-- name: DeleteTag :exec
DELETE FROM tags WHERE id = $1;

-- name: CountSubscribersWithTag :one
SELECT COUNT(*) FROM subscriber_tags WHERE tag_id = $1;

-- name: AddTagToSubscriber :exec
INSERT INTO subscriber_tags (subscriber_id, tag_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveTagFromSubscriber :exec
DELETE FROM subscriber_tags WHERE subscriber_id = $1 AND tag_id = $2;

-- name: IsSubscriberTagged :one
SELECT EXISTS(SELECT 1 FROM subscriber_tags WHERE subscriber_id = $1 AND tag_id = $2);

-- name: ListTagsForSubscriber :many
SELECT t.* FROM tags t
JOIN subscriber_tags st ON st.tag_id = t.id
WHERE st.subscriber_id = $1
ORDER BY t.name ASC;

-- name: ListSubscribersWithTag :many
SELECT s.* FROM subscribers s
JOIN subscriber_tags st ON st.subscriber_id = s.id
WHERE st.tag_id = $1
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

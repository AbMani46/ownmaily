-- name: RecordClick :one
INSERT INTO clicks (campaign_id, subscriber_id, link_index, link_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: CountClicksByCampaign :one
SELECT COUNT(*) FROM clicks WHERE campaign_id = $1;

-- name: CountClicksByLink :many
SELECT link_index, link_url, COUNT(*) AS click_count
FROM clicks
WHERE campaign_id = $1
GROUP BY link_index, link_url
ORDER BY link_index;

-- name: ListClicksBySubscriber :many
SELECT * FROM clicks WHERE subscriber_id = $1 ORDER BY clicked_at DESC;

-- name: RecordOpen :exec
INSERT INTO opens (campaign_id, subscriber_id)
VALUES ($1, $2)
ON CONFLICT (campaign_id, subscriber_id) DO NOTHING;

-- name: CountOpensByCampaign :one
SELECT COUNT(*) FROM opens WHERE campaign_id = $1;

-- name: HasOpened :one
SELECT EXISTS(
    SELECT 1 FROM opens WHERE campaign_id = $1 AND subscriber_id = $2
);

-- name: ListOpensBySubscriber :many
SELECT * FROM opens WHERE subscriber_id = $1 ORDER BY opened_at DESC;

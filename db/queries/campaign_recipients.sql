-- name: CreateCampaignRecipient :one
INSERT INTO campaign_recipients (campaign_id, subscriber_id)
VALUES ($1, $2)
RETURNING *;

-- name: BulkCreateCampaignRecipients :copyfrom
INSERT INTO campaign_recipients (campaign_id, subscriber_id) VALUES ($1, $2);

-- name: GetRecipientStatus :one
SELECT status FROM campaign_recipients
WHERE campaign_id = $1 AND subscriber_id = $2;

-- name: UpdateRecipientStatus :exec
UPDATE campaign_recipients SET status = $3, sent_at = $4
WHERE campaign_id = $1 AND subscriber_id = $2;

-- name: ListPendingRecipients :many
SELECT * FROM campaign_recipients
WHERE campaign_id = $1 AND status = 'pending'
ORDER BY id
LIMIT $2 OFFSET $3;

-- name: CountRecipientsByStatus :one
SELECT COUNT(*) FROM campaign_recipients
WHERE campaign_id = $1 AND status = $2;

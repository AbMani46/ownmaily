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

-- name: CountTotalSent :one
SELECT COUNT(*) FROM campaign_recipients WHERE status = 'sent';

-- name: ListCampaignsReceivedBySubscriber :many
SELECT c.id, c.name, c.subject, c.preview_text, c.from_name, c.from_email, c.reply_to,
       c.html_body, c.text_body, c.status, c.send_to_type, c.send_to_id,
       c.scheduled_at, c.sent_at, c.created_at, c.updated_at
FROM campaigns c
JOIN campaign_recipients cr ON cr.campaign_id = c.id
WHERE cr.subscriber_id = $1
ORDER BY cr.sent_at DESC NULLS LAST;

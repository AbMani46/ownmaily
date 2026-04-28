-- name: CreateCampaign :one
INSERT INTO campaigns (name, subject, preview_text, from_name, from_email, reply_to, html_body, text_body, send_to_type, send_to_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetCampaignByID :one
SELECT * FROM campaigns WHERE id = $1;

-- name: ListCampaigns :many
SELECT * FROM campaigns ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: CountCampaigns :one
SELECT COUNT(*) FROM campaigns;

-- name: UpdateCampaign :one
UPDATE campaigns SET
    name         = $2,
    subject      = $3,
    preview_text = $4,
    from_name    = $5,
    from_email   = $6,
    reply_to     = $7,
    html_body    = $8,
    text_body    = $9,
    send_to_type = $10,
    send_to_id   = $11,
    scheduled_at = $12,
    updated_at   = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCampaignStatus :exec
UPDATE campaigns SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: DeleteCampaign :exec
DELETE FROM campaigns WHERE id = $1;

-- name: ListScheduledCampaignsDue :many
SELECT * FROM campaigns
WHERE status = 'scheduled' AND scheduled_at <= NOW();

-- name: ListCampaignsByStatus :many
SELECT * FROM campaigns
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountCampaignsByStatus :one
SELECT COUNT(*) FROM campaigns WHERE status = $1;

-- name: ScheduleCampaign :one
UPDATE campaigns
SET status = 'scheduled', scheduled_at = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CancelCampaign :one
UPDATE campaigns
SET status = 'draft', scheduled_at = NULL, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkCampaignSent :exec
UPDATE campaigns
SET status = 'sent', sent_at = NOW(), updated_at = NOW()
WHERE id = $1;

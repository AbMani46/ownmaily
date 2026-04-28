-- name: CreateSendJob :one
INSERT INTO send_jobs (campaign_id) VALUES ($1) RETURNING *;

-- name: GetSendJob :one
SELECT * FROM send_jobs WHERE id = $1;

-- name: GetSendJobByCampaign :one
SELECT * FROM send_jobs WHERE campaign_id = $1 ORDER BY created_at DESC LIMIT 1;

-- name: UpdateSendJobStatus :exec
UPDATE send_jobs SET status = $2, updated_at = NOW() WHERE id = $1;

-- name: IncrementSentCount :exec
UPDATE send_jobs SET sent_count = sent_count + 1, updated_at = NOW() WHERE id = $1;

-- name: IncrementFailedCount :exec
UPDATE send_jobs SET failed_count = failed_count + 1, updated_at = NOW() WHERE id = $1;

-- name: ListPendingSendJobs :many
SELECT * FROM send_jobs WHERE status = 'pending' ORDER BY created_at ASC;

-- name: UpdateSendJobCounts :exec
UPDATE send_jobs
SET total_count = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateSendJobError :exec
UPDATE send_jobs
SET status = $2, error_message = $3, updated_at = NOW()
WHERE id = $1;

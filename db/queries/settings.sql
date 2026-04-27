-- name: GetSettings :one
SELECT * FROM settings WHERE id = TRUE;

-- name: UpdateSettings :one
UPDATE settings SET
    site_name        = $1,
    installation_url = $2,
    timezone         = $3,
    physical_address = $4,
    from_name        = $5,
    from_email       = $6,
    reply_to         = $7,
    smtp_provider    = $8,
    smtp_credentials = $9,
    updated_at       = NOW()
WHERE id = TRUE
RETURNING *;

-- name: SetSetupComplete :exec
UPDATE settings SET setup_complete = TRUE, updated_at = NOW() WHERE id = TRUE;

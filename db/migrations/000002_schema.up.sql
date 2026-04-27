-- settings: single row, always exists after setup wizard
CREATE TABLE settings (
    id               BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id = TRUE),
    setup_complete   BOOLEAN NOT NULL DEFAULT FALSE,
    site_name        TEXT NOT NULL DEFAULT 'OwnMaily',
    installation_url TEXT NOT NULL DEFAULT '',
    timezone         TEXT NOT NULL DEFAULT 'UTC',
    physical_address TEXT NOT NULL DEFAULT '',
    from_name        TEXT NOT NULL DEFAULT '',
    from_email       TEXT NOT NULL DEFAULT '',
    reply_to         TEXT NOT NULL DEFAULT '',
    smtp_provider    TEXT NOT NULL DEFAULT '',
    smtp_credentials JSONB NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- owner: single row
CREATE TABLE owner (
    id            BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id = TRUE),
    email         TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- api_keys: one active key at a time
CREATE TABLE api_keys (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_hash   TEXT NOT NULL,
    key_prefix TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- subscribers
CREATE TABLE subscribers (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL DEFAULT '',
    last_name  TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'unsubscribed', 'bounced', 'pending')),
    source     TEXT NOT NULL DEFAULT 'api' CHECK (source IN ('import', 'api', 'form')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- suppression list
CREATE TABLE suppressed_emails (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT NOT NULL UNIQUE,
    reason     TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- lists
CREATE TABLE lists (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL UNIQUE,
    description   TEXT NOT NULL DEFAULT '',
    double_opt_in BOOLEAN NOT NULL DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- list memberships
CREATE TABLE list_subscribers (
    list_id       UUID NOT NULL REFERENCES lists(id) ON DELETE CASCADE,
    subscriber_id UUID NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (list_id, subscriber_id)
);

-- tags
CREATE TABLE tags (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- subscriber tags
CREATE TABLE subscriber_tags (
    subscriber_id UUID NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
    tag_id        UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (subscriber_id, tag_id)
);

-- campaigns
CREATE TABLE campaigns (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL,
    subject      TEXT NOT NULL DEFAULT '',
    preview_text TEXT NOT NULL DEFAULT '',
    from_name    TEXT NOT NULL DEFAULT '',
    from_email   TEXT NOT NULL DEFAULT '',
    reply_to     TEXT NOT NULL DEFAULT '',
    html_body    TEXT NOT NULL DEFAULT '',
    text_body    TEXT NOT NULL DEFAULT '',
    status       TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'scheduled', 'queued', 'sending', 'sent', 'failed')),
    send_to_type TEXT NOT NULL DEFAULT 'list' CHECK (send_to_type IN ('list', 'tag')),
    send_to_id   UUID,
    scheduled_at TIMESTAMPTZ,
    sent_at      TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- send jobs: one row per campaign send
CREATE TABLE send_jobs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id   UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    status        TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'complete', 'failed')),
    total_count   INT NOT NULL DEFAULT 0,
    sent_count    INT NOT NULL DEFAULT 0,
    failed_count  INT NOT NULL DEFAULT 0,
    error_message TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- campaign recipients: one row per subscriber per campaign
CREATE TABLE campaign_recipients (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id   UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    subscriber_id UUID NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
    status        TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed')),
    sent_at       TIMESTAMPTZ,
    UNIQUE (campaign_id, subscriber_id)
);

-- opens: first open only per campaign per subscriber
CREATE TABLE opens (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id   UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    subscriber_id UUID NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
    opened_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (campaign_id, subscriber_id)
);

-- clicks
CREATE TABLE clicks (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id   UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    subscriber_id UUID NOT NULL REFERENCES subscribers(id) ON DELETE CASCADE,
    link_index    INT NOT NULL,
    link_url      TEXT NOT NULL DEFAULT '',
    clicked_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- indexes
CREATE INDEX idx_subscribers_email ON subscribers(email);
CREATE INDEX idx_subscribers_status ON subscribers(status);
CREATE INDEX idx_campaign_recipients_campaign ON campaign_recipients(campaign_id);
CREATE INDEX idx_campaign_recipients_status ON campaign_recipients(status);
CREATE INDEX idx_send_jobs_status ON send_jobs(status);
CREATE INDEX idx_opens_campaign ON opens(campaign_id);
CREATE INDEX idx_clicks_campaign ON clicks(campaign_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_scheduled ON campaigns(scheduled_at) WHERE status = 'scheduled';

-- seed the single settings row
INSERT INTO settings DEFAULT VALUES;

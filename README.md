# OwnMaily

Self-hosted email marketing. No subscriber limits, no monthly fees. Free to self-host -- support development with a one-time $49 payment at [ownmaily.com](https://ownmaily.com).

## What it does

OwnMaily is a self-hosted email marketing tool for indie founders, bootstrappers, and small newsletter operators who want full control over their list without paying $50/month to a SaaS. You run it on your own server, connect your own SMTP provider, and own your data. No subscriber limits, no sending limits, no recurring fees.

## Features

- Send campaigns to unlimited subscribers
- Import and export subscribers via CSV
- Open and click tracking
- Bounce handling (Resend, Mailgun, Amazon SES)
- Scheduled campaigns
- Embeddable signup forms with double opt-in
- Full REST API

## Quick Start

### One-line install (Linux/Mac)

```bash
curl -fsSL https://ownmaily.com/install.sh | bash
```

You will be asked for your installation URL. Everything else is automatic.

[Full installation guide](https://ownmaily.com/docs/getting-started/installation)

### Before going to production

> If you are planning to use OwnMaily seriously, read the documentation at [ownmaily.com/docs](https://ownmaily.com/docs) before getting started. There are a few things to set up -- a domain, SSL, and an SMTP provider -- and the docs walk you through all of it step by step.

Not a reader? Point your AI assistant at https://ownmaily.com/llms.txt and ask it to walk you through setup.

### Manual install (Docker Compose)

```bash
git clone https://github.com/AbMani46/ownmaily
cd ownmaily
cp .env.example .env
docker compose up -d
```

### Windows

Use WSL2 or deploy via Railway or Render.

## SMTP Providers

OwnMaily requires you to bring your own SMTP provider:

| Provider   | Free tier          | Notes                           |
| ---------- | ------------------ | ------------------------------- |
| Resend     | 3,000/month        | Recommended for getting started |
| Mailgun    | Pay as you go      | Good deliverability             |
| Amazon SES | 62,000/month (EC2) | Cheapest at scale               |

Full setup guides at [ownmaily.com/docs/smtp](https://ownmaily.com/docs/smtp)

## Resource Usage

Approximate values for a typical self-hosted install:

| Resource   | Idle                  | Under load          |
| ---------- | --------------------- | ------------------- |
| RAM        | ~50MB                 | ~150MB              |
| Disk (app) | ~30MB                 | ~30MB               |
| Disk (DB)  | ~50MB/10k subscribers | scales linearly     |
| CPU        | negligible            | spikes during sends |

A $6/month VPS (2GB RAM, 20GB disk) is more than enough to get started.

## Updating

```bash
docker compose pull && docker compose up -d
```

Migrations run automatically on startup.

## Other Free Tools

Built by the same developer:

- **Middl** -- [getmiddl.com](https://getmiddl.com), free AI-powered freelance workspace
- **SolidUptime** -- [soliduptime.org](https://soliduptime.org), free uptime monitoring with incident grouping

## License

MIT

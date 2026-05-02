# OwnMaily

Self-hosted email marketing. No subscriber limits, no monthly fees. Free to self-host -- support development with a one-time $49 payment at [ownmaily.com](https://ownmaily.com).

## Features

- Send campaigns to unlimited subscribers
- Import/export CSV
- Open and click tracking
- Bounce handling (Resend, Mailgun, Amazon SES)
- Scheduled campaigns
- Embeddable signup forms
- Double opt-in support
- Full REST API

## Quick Start

### One-line install (Linux/Mac)

```bash
curl -fsSL https://ownmaily.com/install.sh | bash
```

You will be asked for one thing: your installation URL. Everything else is handled automatically -- Docker is installed if missing, a random secret is generated, and the stack starts on port 4400.

Open the URL in your browser to complete setup. The setup wizard will walk you through creating your admin account and connecting an SMTP provider (skippable -- you can add SMTP later in Settings).

### Manual install (Docker Compose)

```bash
git clone https://github.com/AbMani46/ownmaily
cd ownmaily
cp .env.example .env
# Edit .env -- set APP_SECRET and INSTALLATION_URL
docker compose up -d
```

Visit `http://localhost:4400` to complete setup.

### Windows

Windows is not supported by the install script. Use WSL2 or deploy via Railway/Render using the one-click buttons in the repo.

## Resource Usage

Approximate values for a typical self-hosted install:

| Resource   | Idle                  | Under load          |
| ---------- | --------------------- | ------------------- |
| RAM        | ~50MB                 | ~150MB              |
| Disk (app) | ~30MB                 | ~30MB               |
| Disk (DB)  | ~50MB/10k subscribers | scales linearly     |
| CPU        | negligible            | spikes during sends |

A $6/month VPS (2GB RAM, 20GB disk) is more than enough to get started.

## SMTP Providers

OwnMaily requires you to bring your own SMTP provider:

| Provider   | Free tier          | Notes                           |
| ---------- | ------------------ | ------------------------------- |
| Resend     | 3,000/month        | Recommended for getting started |
| Mailgun    | Pay as you go      | Good deliverability             |
| Amazon SES | 62,000/month (EC2) | Cheapest at scale               |

## Updates

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

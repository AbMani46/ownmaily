#!/usr/bin/env bash
set -euo pipefail

# ─── Windows guard ─────────────────────────────────────────────────────────────
if [[ "${OSTYPE:-}" == "msys" || "${OSTYPE:-}" == "cygwin" || "${OS:-}" == "Windows_NT" ]]; then
    echo "Windows is not supported by this installer."
    echo ""
    echo "Options:"
    echo "  - Use WSL2: https://docs.microsoft.com/en-us/windows/wsl/install"
    echo "  - One-click deploy: https://ownmaily.com#deploy"
    exit 1
fi

# ─── Colour helpers ────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${GREEN}[+]${NC} $*"; }
warn()  { echo -e "${YELLOW}[!]${NC} $*"; }
error() { echo -e "${RED}[✗]${NC} $*" >&2; }
die()   { error "$*"; exit 1; }

# Run with privilege: no-op if root, uses sudo otherwise.
_sudo() {
    if [[ $EUID -eq 0 ]]; then "$@"; else sudo "$@"; fi
}

# ─── whiptail: check / install ─────────────────────────────────────────────────
if ! command -v whiptail &>/dev/null; then
    info "whiptail not found — attempting to install..."
    if command -v apt-get &>/dev/null; then
        _sudo apt-get update -qq
        _sudo apt-get install -y -qq whiptail
    elif command -v yum &>/dev/null; then
        _sudo yum install -y newt
    elif command -v brew &>/dev/null; then
        brew install newt
    else
        die "Cannot install whiptail automatically.
Please install it manually:
  Debian/Ubuntu:  sudo apt-get install whiptail
  RHEL/CentOS:    sudo yum install newt
  macOS:          brew install newt
Then re-run this installer."
    fi
fi
command -v whiptail &>/dev/null || die "whiptail installation failed. See above."

# ─── Docker: check / install ───────────────────────────────────────────────────
if ! command -v docker &>/dev/null; then
    info "Docker not found — installing via get.docker.com..."
    if ! curl -fsSL https://get.docker.com | sh; then
        die "Docker installation failed.
Please install Docker manually: https://docs.docker.com/get-docker/
Then re-run this installer."
    fi
    # Add current user to the docker group (takes effect on next login)
    if [[ -n "${USER:-}" ]] && ! id -nG "$USER" 2>/dev/null | grep -qw docker; then
        _sudo usermod -aG docker "$USER" 2>/dev/null || true
        warn "Added $USER to the 'docker' group. You may need to log out and back in."
    fi
fi
command -v docker &>/dev/null || die "Docker is not available after install. Please reboot and re-run."

# ─── Docker Compose v2: check ──────────────────────────────────────────────────
if ! docker compose version &>/dev/null; then
    die "Docker Compose v2 is not available.
Please install it and re-run:
  Ubuntu/Debian:  sudo apt-get install docker-compose-plugin
  Other Linux:    https://docs.docker.com/compose/install/linux/
  Docker Desktop: bundled — install from https://docs.docker.com/get-docker/"
fi

# ─── Welcome screen ────────────────────────────────────────────────────────────
whiptail --title "OwnMaily Installer" \
    --msgbox "\
OwnMaily — Self-Hosted Email Marketing

Own your list. No monthly fees. No limits.

This installer will:
  • Ask for your installation URL and admin account
  • Download and start OwnMaily with Docker
  • Save credentials for the setup wizard

You will need a domain or public IP pointed at this
server, and an SMTP provider (Resend, Mailgun, SES).

Press OK to begin setup." \
    20 62

# ─── Installation URL ──────────────────────────────────────────────────────────
INSTALLATION_URL=""
while true; do
    INSTALLATION_URL=$(whiptail --title "Installation URL" \
        --inputbox "\
The public URL where OwnMaily will be accessed.

This is used for tracking links, unsubscribe URLs,
and email footers — it must be reachable by your subscribers.

Examples:
  https://mail.yourdomain.com
  http://123.45.67.89:4400
" \
        18 66 "https://" 3>&1 1>&2 2>&3) || die "Installation cancelled."

    INSTALLATION_URL="${INSTALLATION_URL%%/}"  # strip trailing slash
    if [[ -z "$INSTALLATION_URL" ]]; then
        whiptail --title "Invalid Input" --msgbox "Installation URL cannot be empty." 8 42
    else
        break
    fi
done

# ─── Admin email ───────────────────────────────────────────────────────────────
ADMIN_EMAIL=""
while true; do
    ADMIN_EMAIL=$(whiptail --title "Admin Account" \
        --inputbox "\
Enter your admin email address.

This will be your login username and the address
test emails are sent to during the setup wizard.
" \
        13 62 "" 3>&1 1>&2 2>&3) || die "Installation cancelled."

    if [[ "$ADMIN_EMAIL" == *"@"* && ${#ADMIN_EMAIL} -gt 3 ]]; then
        break
    fi
    whiptail --title "Invalid Input" \
        --msgbox "Please enter a valid email address (must contain @)." 8 52
done

# ─── Admin password (with confirmation) ────────────────────────────────────────
ADMIN_PASSWORD=""
while true; do
    ADMIN_PASSWORD=$(whiptail --title "Admin Account" \
        --passwordbox "\
Set a password for your admin account.

Minimum 8 characters.
" \
        11 55 "" 3>&1 1>&2 2>&3) || die "Installation cancelled."

    if [[ ${#ADMIN_PASSWORD} -lt 8 ]]; then
        whiptail --title "Invalid Input" \
            --msgbox "Password must be at least 8 characters. Please try again." 8 55
        continue
    fi

    ADMIN_PASSWORD_CONFIRM=$(whiptail --title "Admin Account" \
        --passwordbox "Confirm your password." \
        8 45 "" 3>&1 1>&2 2>&3) || die "Installation cancelled."

    if [[ "$ADMIN_PASSWORD" == "$ADMIN_PASSWORD_CONFIRM" ]]; then
        break
    fi
    whiptail --title "Passwords Do Not Match" \
        --msgbox "The passwords you entered do not match. Please try again." 8 55
done

# ─── Generate secrets ──────────────────────────────────────────────────────────
APP_SECRET=$(openssl rand -hex 32)
POSTGRES_PASSWORD=$(openssl rand -hex 16)

# ─── Create install directory ──────────────────────────────────────────────────
INSTALL_DIR="$HOME/ownmaily"
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"
info "Installation directory: $INSTALL_DIR"

# ─── Download docker-compose.yml ───────────────────────────────────────────────
info "Downloading docker-compose.yml..."
if ! curl -fsSL \
    https://raw.githubusercontent.com/AbMani46/ownmaily/main/docker-compose.yml \
    -o docker-compose.yml; then
    die "Failed to download docker-compose.yml.
Check your internet connection and try again."
fi

# ─── Write .env ────────────────────────────────────────────────────────────────
cat > .env <<EOF
APP_SECRET=${APP_SECRET}
INSTALLATION_URL=${INSTALLATION_URL}
PORT=4400
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
EOF
chmod 600 .env
info "Created .env"

# ─── Write setup credentials ───────────────────────────────────────────────────
cat > setup_credentials.txt <<EOF
OwnMaily Setup Credentials
==========================
Admin Email:    ${ADMIN_EMAIL}
Admin Password: ${ADMIN_PASSWORD}

These credentials are required to complete the Setup Wizard on first launch.
Have them ready when you visit ${INSTALLATION_URL}.

Delete this file after your setup wizard is complete.
EOF
chmod 600 setup_credentials.txt
info "Credentials saved to $INSTALL_DIR/setup_credentials.txt"
warn "Delete setup_credentials.txt after completing the setup wizard."

# ─── Pull images and start ─────────────────────────────────────────────────────
info "Pulling Docker images — this may take a minute..."
docker compose pull

info "Starting OwnMaily..."
docker compose up -d

# ─── Health check ──────────────────────────────────────────────────────────────
HEALTH_URL="http://localhost:4400/health"
TIMEOUT=60
ELAPSED=0
READY=0

info "Waiting for OwnMaily to become ready..."
while [[ $ELAPSED -lt $TIMEOUT ]]; do
    if curl -sf "$HEALTH_URL" &>/dev/null; then
        READY=1
        break
    fi
    printf "."
    sleep 3
    ELAPSED=$((ELAPSED + 3))
done
echo ""  # newline after dots

if [[ $READY -ne 1 ]]; then
    warn "OwnMaily did not respond at $HEALTH_URL within ${TIMEOUT}s."
    warn "It may still be starting. Check logs with:"
    warn "  docker compose -f $INSTALL_DIR/docker-compose.yml logs -f"
fi

# ─── Success screen ────────────────────────────────────────────────────────────
if [[ $READY -eq 1 ]]; then
    whiptail --title "OwnMaily is Ready" \
        --msgbox "\
OwnMaily is running!

Visit your installation URL to complete setup:
  ${INSTALLATION_URL}

Your admin email is:
  ${ADMIN_EMAIL}

Your credentials are saved in:
  ${INSTALL_DIR}/setup_credentials.txt

Delete that file once your setup wizard is complete." \
        20 66
else
    whiptail --title "OwnMaily Started" \
        --msgbox "\
OwnMaily has started, but the health check timed out.

The app may still be initialising. Try visiting:
  ${INSTALLATION_URL}

Your admin email is:
  ${ADMIN_EMAIL}

If the site does not load, check the logs:
  docker compose -f ${INSTALL_DIR}/docker-compose.yml logs -f" \
        20 66
fi

# ─── Final stdout summary ──────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
if [[ $READY -eq 1 ]]; then
    echo -e "${GREEN}  OwnMaily is running${NC}"
else
    echo -e "${YELLOW}  OwnMaily started (health check timed out)${NC}"
fi
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "  URL:         ${BOLD}${INSTALLATION_URL}${NC}"
echo -e "  Admin email: ${BOLD}${ADMIN_EMAIL}${NC}"
echo -e "  Credentials: ${BOLD}${INSTALL_DIR}/setup_credentials.txt${NC}"
echo ""
echo "  Next: visit your URL and complete the setup wizard."
echo "  Then delete setup_credentials.txt."
echo ""
echo "  Useful commands:"
echo "    docker compose -f ${INSTALL_DIR}/docker-compose.yml logs -f"
echo "    docker compose -f ${INSTALL_DIR}/docker-compose.yml restart"
echo "    docker compose -f ${INSTALL_DIR}/docker-compose.yml down"
echo ""

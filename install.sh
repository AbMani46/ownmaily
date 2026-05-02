#!/usr/bin/env bash
set -euo pipefail

# ── Windows guard ──────────────────────────────────────────────────────────────
if [[ "${OSTYPE:-}" == "msys" || "${OSTYPE:-}" == "cygwin" || "${OS:-}" == "Windows_NT" ]]; then
    echo "Windows is not supported by this installer."
    echo ""
    echo "Options:"
    echo "  WSL2: https://docs.microsoft.com/en-us/windows/wsl/install"
    echo "  One-click deploy: https://ownmaily.com#deploy"
    exit 1
fi

# ── Colors ─────────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'
NC='\033[0m'

info()  { echo -e "${GREEN}[+]${NC} $*"; }
warn()  { echo -e "${YELLOW}[!]${NC} $*"; }
error() { echo -e "${RED}[x]${NC} $*" >&2; }
die()   { error "$*"; exit 1; }

# No-op if root, sudo otherwise.
_sudo() {
    if [[ $EUID -eq 0 ]]; then "$@"; else sudo "$@"; fi
}

# ── Docker: check / install ────────────────────────────────────────────────────
if ! command -v docker &>/dev/null; then
    info "Docker not found. Installing via get.docker.com..."
    if ! curl -fsSL https://get.docker.com | sh; then
        die "Docker installation failed. Install manually: https://docs.docker.com/get-docker/"
    fi
    if [[ -n "${USER:-}" ]] && ! id -nG "$USER" 2>/dev/null | grep -qw docker; then
        _sudo usermod -aG docker "$USER" 2>/dev/null || true
        warn "Added $USER to the docker group. You may need to log out and back in."
    fi
fi
command -v docker &>/dev/null || die "Docker is not available after install. Please reboot and re-run."

# ── Docker Compose v2: check ───────────────────────────────────────────────────
if ! docker compose version &>/dev/null; then
    die "Docker Compose v2 is not available. Install it and re-run:
  Ubuntu/Debian:  sudo apt-get install docker-compose-plugin
  Other Linux:    https://docs.docker.com/compose/install/linux/
  Docker Desktop: https://docs.docker.com/get-docker/"
fi

# ── Idempotency check ─────────────────────────────────────────────────────────
INSTALL_DIR="$HOME/ownmaily"

if [ -d "$INSTALL_DIR" ] && [ -f "$INSTALL_DIR/.env" ]; then
    echo "" >/dev/tty
    echo "[!] OwnMaily is already installed at $INSTALL_DIR" >/dev/tty
    echo "" >/dev/tty
    echo "    To update:    cd $INSTALL_DIR && docker compose pull && docker compose up -d" >/dev/tty
    echo "    To restart:   cd $INSTALL_DIR && docker compose restart" >/dev/tty
    echo "    To uninstall: cd $INSTALL_DIR && docker compose down -v" >/dev/tty
    echo "" >/dev/tty
    read -r -p "Run anyway and overwrite existing install? [y/N] " confirm </dev/tty
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "Exiting. Existing install unchanged."
        exit 0
    fi
    info "Removing existing install..."
    cd "$INSTALL_DIR"
    docker compose down -v 2>/dev/null || true
    cd "$HOME"
fi

# ── Prompts (all via /dev/tty so curl-pipe works) ─────────────────────────────
echo "" >/dev/tty
echo -e "${BOLD}OwnMaily Setup${NC}" >/dev/tty
echo "Answer one question and the rest is automatic." >/dev/tty
echo "" >/dev/tty

# Installation URL
echo "  What to enter here:" >/dev/tty
echo "" >/dev/tty
echo "  Testing locally:      http://localhost:4400" >/dev/tty
echo "  On a remote server:   http://$(curl -sf https://ifconfig.me || echo '<your-server-ip>'):4400" >/dev/tty
echo "  With a domain:        https://mail.yourdomain.com" >/dev/tty
echo "" >/dev/tty
echo "  Not sure? Use the remote server line above -- your IP was auto-detected." >/dev/tty
echo "" >/dev/tty
printf "${BOLD}Installation URL${NC} [http://localhost:4400]: " >/dev/tty
read -r INSTALLATION_URL </dev/tty
INSTALLATION_URL="${INSTALLATION_URL:-http://localhost:4400}"
INSTALLATION_URL="${INSTALLATION_URL%%/}"

echo "" >/dev/tty

# ── Generate secrets ───────────────────────────────────────────────────────────
APP_SECRET=$(openssl rand -hex 32)
POSTGRES_PASSWORD=ownmaily
DB_URL="postgres://ownmaily:${POSTGRES_PASSWORD}@db:5432/ownmaily?sslmode=disable"

# ── Create install directory ───────────────────────────────────────────────────
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"
info "Install directory: $INSTALL_DIR"

# ── Download docker-compose.yml ────────────────────────────────────────────────
info "Downloading docker-compose.yml..."
if ! curl -fsSL \
    https://raw.githubusercontent.com/AbMani46/ownmaily/main/docker-compose.yml \
    -o docker-compose.yml; then
    die "Failed to download docker-compose.yml. Check your internet connection and try again."
fi

# ── Write .env ─────────────────────────────────────────────────────────────────
cat > .env <<EOF
APP_SECRET=${APP_SECRET}
DB_URL=${DB_URL}
INSTALLATION_URL=${INSTALLATION_URL}
PORT=4400
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
EOF
chmod 600 .env
info "Created .env"

# ── Pull images and start ──────────────────────────────────────────────────────
info "Pulling Docker images (this may take a minute)..."
docker compose pull

info "Starting OwnMaily..."
docker compose up -d

# ── Health check ───────────────────────────────────────────────────────────────
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
echo ""

# ── Success summary ────────────────────────────────────────────────────────────
echo ""
echo -e "${BOLD}============================================${NC}"
if [[ $READY -eq 1 ]]; then
    echo -e "${GREEN}  OwnMaily is running${NC}"
else
    echo -e "${YELLOW}  OwnMaily started (health check timed out)${NC}"
fi
echo -e "${BOLD}============================================${NC}"
echo ""
echo -e "  OwnMaily is running at ${BOLD}${INSTALLATION_URL}${NC}"
echo -e "  Open that URL in your browser to complete setup."
echo ""
if [[ $READY -ne 1 ]]; then
    warn "Health check timed out. OwnMaily may still be starting up."
    warn "Check logs: docker compose -f ${INSTALL_DIR}/docker-compose.yml logs -f"
    echo ""
fi

# ── Self-cleanup (manual run only) ────────────────────────────────────────────
if [[ -f "$0" && "$0" != "bash" && "$0" != "/dev/stdin" ]]; then
    rm -f "$0" 2>/dev/null || true
fi

#!/bin/bash
set -e

echo "Installing OwnMaily..."

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "Docker not found. Installing..."
    curl -fsSL https://get.docker.com | sh
fi

# Check Docker Compose
if ! docker compose version &> /dev/null; then
    echo "Docker Compose v2 not found. Please install Docker Desktop or Docker Compose v2."
    exit 1
fi

# Create directory
mkdir -p ownmaily && cd ownmaily

# Download docker-compose.yml
curl -fsSL https://raw.githubusercontent.com/AbMani46/ownmaily/main/docker-compose.yml -o docker-compose.yml

# Generate .env
if [ ! -f .env ]; then
    SECRET=$(openssl rand -hex 32 2>/dev/null || cat /dev/urandom | tr -dc 'a-f0-9' | head -c 64)
    cat > .env << EOF
APP_SECRET=${SECRET}
INSTALLATION_URL=http://localhost:4400
PORT=4400
POSTGRES_PASSWORD=ownmaily
EOF
    echo ".env created. Edit INSTALLATION_URL to your public domain before going live."
fi

# Start
docker compose pull
docker compose up -d

echo ""
echo "OwnMaily is running at http://localhost:4400"
echo "Visit that URL to complete setup."

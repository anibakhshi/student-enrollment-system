#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIRECTORY/.." && pwd)"

cd "$PROJECT_ROOT"

if [ ! -f ".env" ]; then
    echo "Error: .env file does not exist."
    echo "Run: cp .env.example .env"
    exit 1
fi

echo "Stopping the containerized API to release port 8081..."
docker compose stop api >/dev/null 2>&1 || true

echo "Starting PostgreSQL and Adminer..."
docker compose up -d postgres adminer

echo "Running database migrations..."
docker compose run --rm migrate

echo "Loading development environment variables..."
set -a
source .env
set +a

echo "Starting Go API in development mode..."
echo "Press Ctrl+C to stop the API."

exec go run ./cmd/api
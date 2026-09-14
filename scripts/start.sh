#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIRECTORY/.." && pwd)"

cd "$PROJECT_ROOT"

if [ ! -f ".env" ]; then
    echo "Error: .env file does not exist."
    echo "Create it from .env.example:"
    echo "cp .env.example .env"
    exit 1
fi

echo "Validating Docker Compose configuration..."
docker compose config --quiet

echo "Building and starting the application..."
docker compose up -d --build

echo "Waiting for the API..."

for attempt in $(seq 1 30); do
    if curl --silent --fail http://localhost:8081/health >/dev/null; then
        echo "API is ready: http://localhost:8081"
        docker compose ps
        exit 0
    fi

    sleep 1
done

echo "Error: API did not become ready."
docker compose ps -a
docker compose logs --tail=100 api
exit 1
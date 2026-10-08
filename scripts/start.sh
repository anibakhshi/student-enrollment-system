#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIRECTORY="$(
    cd "$(dirname "${BASH_SOURCE[0]}")" &&
    pwd
)"

PROJECT_ROOT="$(
    cd "$SCRIPT_DIRECTORY/.." &&
    pwd
)"

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

echo "Waiting for the Go REST API..."

api_ready=false

for attempt in $(seq 1 60); do
    if curl \
        --silent \
        --fail \
        http://localhost:8081/health \
        >/dev/null; then
        api_ready=true
        break
    fi

    sleep 2
done

if [ "$api_ready" != "true" ]; then
    echo "Error: Go REST API did not become ready."
    docker compose ps -a
    docker compose logs --tail=100 api
    exit 1
fi

echo "Go REST API is ready:"
echo "http://localhost:8081"

echo "Waiting for the Django GUI..."

gui_ready=false

for attempt in $(seq 1 180); do
    if curl \
        --silent \
        --fail \
        http://localhost:8000/ \
        >/dev/null; then
        gui_ready=true
        break
    fi

    sleep 2
done

if [ "$gui_ready" != "true" ]; then
    echo "Error: Django GUI did not become ready."
    docker compose ps -a
    docker compose logs --tail=150 gui
    exit 1
fi

echo "Django GUI is ready:"
echo "http://localhost:8000"

echo "Adminer is available:"
echo "http://localhost:18080"

docker compose ps
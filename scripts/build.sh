#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIRECTORY/.." && pwd)"

cd "$PROJECT_ROOT"

echo "Validating Docker Compose configuration..."
docker compose config --quiet

echo "Building Go API image..."
docker compose --progress plain build api

echo "Build completed successfully."

docker images student-enrollment-api:local
#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIRECTORY/.." && pwd)"

cd "$PROJECT_ROOT"

echo "Stopping application containers..."
docker compose down

echo "Application stopped."
echo "Database data volume was preserved."
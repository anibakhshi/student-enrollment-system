#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIRECTORY/.." && pwd)"

cd "$PROJECT_ROOT"

SERVICE_NAME="${1:-api}"
case "$SERVICE_NAME" in
    api|gui|postgres|migrate|adminer)
        ;;
    *)
        echo "Error: unknown service: $SERVICE_NAME"
        echo "Available services: api, gui, postgres, migrate, adminer"
        exit 1
        ;;
esac

echo "Showing logs for service: $SERVICE_NAME"
echo "Press Ctrl+C to exit."

docker compose logs \
    --follow \
    --tail=100 \
    "$SERVICE_NAME"
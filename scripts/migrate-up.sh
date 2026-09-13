#!/usr/bin/env bash

set -Eeuo pipefail

CONTAINER_NAME="student_enrollment_postgres"
DATABASE_NAME="${POSTGRES_DB:-student_enrollment}"
DATABASE_USER="${POSTGRES_USER:-student_app}"
MIGRATION_FILE="migrations/000001_create_students.up.sql"

if ! docker inspect "$CONTAINER_NAME" >/dev/null 2>&1; then
    echo "Error: PostgreSQL container does not exist."
    echo "Run: docker compose up -d postgres"
    exit 1
fi

if [ "$(docker inspect -f '{{.State.Health.Status}}' "$CONTAINER_NAME")" != "healthy" ]; then
    echo "Error: PostgreSQL container is not healthy."
    exit 1
fi

if [ ! -f "$MIGRATION_FILE" ]; then
    echo "Error: Migration file not found: $MIGRATION_FILE"
    exit 1
fi

echo "Applying database migrations..."

docker exec -i "$CONTAINER_NAME" \
    psql \
    --username "$DATABASE_USER" \
    --dbname "$DATABASE_NAME" \
    --set ON_ERROR_STOP=1 \
    < "$MIGRATION_FILE"

echo "Database migrations applied successfully."
#!/usr/bin/env bash

set -Eeuo pipefail

CONTAINER_NAME="student_enrollment_postgres"
DATABASE_NAME="${POSTGRES_DB:-student_enrollment}"
DATABASE_USER="${POSTGRES_USER:-student_app}"
MIGRATION_FILE="migrations/000001_create_students.down.sql"

if ! docker inspect "$CONTAINER_NAME" >/dev/null 2>&1; then
    echo "Error: PostgreSQL container does not exist."
    exit 1
fi

if [ ! -f "$MIGRATION_FILE" ]; then
    echo "Error: Migration file not found: $MIGRATION_FILE"
    exit 1
fi

echo "Warning: this operation removes the students table."

if [ "${CONFIRM_MIGRATION_DOWN:-}" != "yes" ]; then
    echo "Migration cancelled."
    echo "To confirm, run:"
    echo "CONFIRM_MIGRATION_DOWN=yes ./scripts/migrate-down.sh"
    exit 1
fi

docker exec -i "$CONTAINER_NAME" \
    psql \
    --username "$DATABASE_USER" \
    --dbname "$DATABASE_NAME" \
    --set ON_ERROR_STOP=1 \
    < "$MIGRATION_FILE"

echo "Database migration rolled back successfully."
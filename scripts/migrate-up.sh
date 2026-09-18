#!/usr/bin/env bash

set -Eeuo pipefail

CONTAINER_NAME="student_enrollment_postgres"
DATABASE_NAME="${POSTGRES_DB:-student_enrollment}"
DATABASE_USER="${POSTGRES_USER:-student_app}"
MIGRATIONS_DIRECTORY="migrations"

if ! docker inspect "$CONTAINER_NAME" >/dev/null 2>&1; then
    echo "Error: PostgreSQL container does not exist."
    echo "Run: docker compose up -d postgres"
    exit 1
fi

if [ "$(docker inspect -f '{{.State.Health.Status}}' "$CONTAINER_NAME")" != "healthy" ]; then
    echo "Error: PostgreSQL container is not healthy."
    exit 1
fi

if [ ! -d "$MIGRATIONS_DIRECTORY" ]; then
    echo "Error: migrations directory does not exist."
    exit 1
fi

shopt -s nullglob

migration_files=(
    "$MIGRATIONS_DIRECTORY"/*.up.sql
)

if [ "${#migration_files[@]}" -eq 0 ]; then
    echo "No migration files found."
    exit 0
fi

echo "Checking database migrations..."

for migration_file in "${migration_files[@]}"; do
    migration_name="$(basename "$migration_file")"
    migration_version="${migration_name%%_*}"

    migration_table_exists="$(
        docker exec "$CONTAINER_NAME" \
            psql \
            --username "$DATABASE_USER" \
            --dbname "$DATABASE_NAME" \
            --tuples-only \
            --no-align \
            --command \
            "SELECT to_regclass('public.schema_migrations') IS NOT NULL;"
    )"

    if [ "$migration_table_exists" = "t" ]; then
        migration_applied="$(
            docker exec "$CONTAINER_NAME" \
                psql \
                --username "$DATABASE_USER" \
                --dbname "$DATABASE_NAME" \
                --tuples-only \
                --no-align \
                --command \
                "SELECT EXISTS (
                    SELECT 1
                    FROM schema_migrations
                    WHERE version = '$migration_version'
                );"
        )"

        if [ "$migration_applied" = "t" ]; then
            echo "Skipping $migration_name (already applied)."
            continue
        fi
    fi

    echo "Applying $migration_name..."

    docker exec -i "$CONTAINER_NAME" \
        psql \
        --username "$DATABASE_USER" \
        --dbname "$DATABASE_NAME" \
        --set ON_ERROR_STOP=1 \
        < "$migration_file"

    echo "Applied $migration_name successfully."
done

echo "All database migrations are up to date."
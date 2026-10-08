#!/bin/sh

set -eu

cd /app

echo "Applying Django database migrations..."
python manage.py migrate --noinput

echo "Loading initial course catalog..."
python manage.py seed_catalog

sync_enabled="${DJANGO_SYNC_SEMATEC_ON_START:-true}"

sync_marker="/app/data/.sematec-catalog-synced-v1"

if [ "$sync_enabled" = "true" ] &&
   [ ! -f "$sync_marker" ]; then
    echo "Synchronizing Sematec catalog..."

    if python manage.py sync_sematec_catalog \
        --delay "${SEMATEC_SYNC_DELAY:-0.2}"; then
        touch "$sync_marker"
        echo "Sematec catalog synchronized successfully."
    else
        echo "Warning: Sematec synchronization failed."
        echo "Django will continue with seeded catalog data."
    fi
else
    echo "Sematec catalog synchronization skipped."
fi

echo "Collecting Django static files..."
python manage.py collectstatic \
    --noinput \
    --clear

echo "Starting Django GUI..."
exec "$@"
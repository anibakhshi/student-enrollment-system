#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIRECTORY="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIRECTORY/.." && pwd)"

cd "$PROJECT_ROOT"

mapfile -t GO_FILES < <(
    find . \
        -type f \
        -name "*.go" \
        -not -path "./vendor/*" \
        | sort
)

echo "Checking Go formatting..."

UNFORMATTED_FILES="$(gofmt -l "${GO_FILES[@]}")"

if [ -n "$UNFORMATTED_FILES" ]; then
    echo "Error: these files are not formatted:"
    echo "$UNFORMATTED_FILES"
    echo "Run: go fmt ./..."
    exit 1
fi

echo "Running go vet..."
go vet ./...

echo "Running unit and integration tests..."
go test -v ./...

echo "Running race detector..."
go test -race ./...

echo "All checks passed successfully."
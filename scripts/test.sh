#!/usr/bin/env bash
set -e

# Get the absolute path to the root directory
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." >/dev/null 2>&1 && pwd)"

echo "Running tests..."
cd "$DIR/src" || exit 1
go test -v ./internal/services ./internal/handlers

echo ""
echo "✅ All tests passed!"

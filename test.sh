#!/bin/bash
set -e

echo "Running tests..."
go test -v ./internal/services ./internal/handlers

echo ""
echo "✅ All tests passed!"

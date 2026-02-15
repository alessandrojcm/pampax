#!/usr/bin/env bash

# Stage 0.2 Fixture Generation Script
# Runs inside Docker to ensure clean, reproducible environment

set -e

echo "=========================================="
echo "PAMPAX Stage 0.2 - Fixture Generation"
echo "=========================================="
echo ""

# Change to the fixtures directory
cd "$(dirname "$0")"

# Check if Docker is available
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed or not in PATH"
    exit 1
fi

# Build Docker image
echo "🐳 Building Docker image..."
docker-compose build

echo ""
echo "✅ Docker image built successfully!"
echo ""
echo "You can now generate fixtures using:"
echo ""
echo "  docker-compose run --rm pampax-fixture-gen node test/fixtures/generate-fixtures.js <repo-path> <size>"
echo ""
echo "Or start an interactive shell:"
echo ""
echo "  docker-compose run --rm pampax-fixture-gen /bin/bash"
echo ""
echo "Example workflow:"
echo "  1. docker-compose run --rm pampax-fixture-gen /bin/bash"
echo "  2. Inside container: node test/fixtures/generate-fixtures.js /pampax small"
echo "  3. Fixtures will be saved to test/fixtures/small/"
echo ""
